extern crate alloc;
extern crate wee_alloc;

use std::{io::Cursor, mem::{ManuallyDrop, MaybeUninit}, slice, u32};
use alloc::vec::Vec;
use image::{codecs::jpeg::JpegEncoder, io::Reader as ImageReader, ImageEncoder};

fn thumbnail(img_bytes: &[u8], width: u32, height: u32) -> Vec<u8> {
    let mut res: Vec<u8> = Vec::new();
    let gfb = ImageReader::new(Cursor::new(img_bytes)).with_guessed_format();
    if gfb.is_err() {
        return res;
    }
    let dy_img_res = gfb.unwrap().decode();
    if dy_img_res.is_err() {
        return res;
    }
    let img_buf = dy_img_res.unwrap().into_rgb8();
    let thumbnail = image::imageops::thumbnail(&img_buf, width, height);
    let enc = JpegEncoder::new(&mut res).write_image(&thumbnail, width, height, image::ExtendedColorType::Rgb8);
    if enc.is_err() {
        return res;
    }
    return res;
}

#[cfg_attr(all(target_arch = "wasm32"), export_name = "thumbnail")]
#[no_mangle]
pub unsafe extern "C" fn _thumbnail(ptr: u32, len: u32, width: u32, height: u32) -> u64 {
    let binding = ptr_to_u8_vec(ptr, len);
    let res = thumbnail(binding.as_slice(), width, height);
    let mut v = ManuallyDrop::new(res);
    let (ptr_res, len_res) = (v.as_mut_ptr(), v.len());
    return ((ptr_res as u64) << 32) | len_res as u64;
}

unsafe fn ptr_to_u8_vec(ptr: u32, len: u32) -> Vec<u8> {
    return slice::from_raw_parts(ptr as *mut u8, len as usize).to_vec();
}

/// Returns a pointer and size pair for the given u8 slice in a way compatible
/// with WebAssembly numeric types.
///
/// Note: This doesn't change the ownership of the String. To intentionally
/// leak it, use [`std::mem::forget`] on the input after calling this.
unsafe fn u8_slice_to_ptr(b : &[u8]) -> (u32, u32) {
    return (b.as_ptr() as u32, b.len() as u32);
}

/// Set the global allocator to the WebAssembly optimized one.
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

/// WebAssembly export that allocates a pointer (linear memory offset) that can
/// be used for a string.
///
/// This is an ownership transfer, which means the caller must call
/// [`deallocate`] when finished.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "allocate")]
#[no_mangle]
pub extern "C" fn _allocate(size: u32) -> *mut u8 {
    allocate(size as usize)
}

/// Allocates size bytes and leaks the pointer where they start.
fn allocate(size: usize) -> *mut u8 {
    // Allocate the amount of bytes needed.
    let vec: Vec<MaybeUninit<u8>> = Vec::with_capacity(size);

    // into_raw leaks the memory to the caller.
    Box::into_raw(vec.into_boxed_slice()) as *mut u8
}


/// WebAssembly export that deallocates a pointer of the given size (linear
/// memory offset, byteCount) allocated by [`allocate`].
#[cfg_attr(all(target_arch = "wasm32"), export_name = "deallocate")]
#[no_mangle]
pub unsafe extern "C" fn _deallocate(ptr: u32, size: u32) {
    deallocate(ptr as *mut u8, size as usize);
}

/// Retakes the pointer which allows its memory to be freed.
unsafe fn deallocate(ptr: *mut u8, size: usize) {
    let _ = Vec::from_raw_parts(ptr, 0, size);
}
