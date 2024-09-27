use std::{io::Cursor, mem::ManuallyDrop, slice, u32};
use image::{codecs::jpeg::JpegEncoder, io::Reader as ImageReader, ImageEncoder};
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
extern {
}

#[wasm_bindgen]
pub fn thumbnail_js(img_bytes: &[u8], width: u32, height: u32) -> js_sys::Uint8Array {
    return js_sys::Uint8Array::from(thumbnail(img_bytes, width, height).as_slice());
}

pub fn thumbnail(img_bytes: &[u8], width: u32, height: u32) -> Vec<u8> {
    let image_rs: Result<Vec<u8>, image::ImageError> = thumbnail_image_rs(img_bytes, width, height);
    if !image_rs.is_err() {
        return image_rs.unwrap();
    }
    vec![]
}

pub fn thumbnail_image_rs(img_bytes: &[u8], width: u32, height: u32) -> Result<Vec<u8>, image::ImageError> {
    let mut res: Vec<u8> = Vec::new();
    let img_buf: image::ImageBuffer<image::Rgb<u8>, Vec<u8>>= ImageReader::new(Cursor::new(img_bytes)).with_guessed_format()?.decode()?.into_rgb8();
    let thumbnail: image::ImageBuffer<image::Rgb<u8>, Vec<u8>> = image::imageops::thumbnail(&img_buf, width, height);
    JpegEncoder::new(&mut res).write_image(&thumbnail, width, height, image::ExtendedColorType::Rgb8)?;
    return Result::Ok(res);
}

#[cfg_attr(all(target_arch = "wasm32"), export_name = "thumbnail")]
#[no_mangle]
pub unsafe extern "C" fn _thumbnail_ptr(ptr: u32, len: u32, width: u32, height: u32) -> u64 {
    let binding: Vec<u8> = ptr_to_u8_vec(ptr, len);
    let res: Vec<u8> = thumbnail(binding.as_slice(), width, height);
    let mut v: ManuallyDrop<Vec<u8>> = ManuallyDrop::new(res);
    let (ptr_res, len_res) = (v.as_mut_ptr(), v.len());
    return ((ptr_res as u64) << 32) | len_res as u64;
}

unsafe fn ptr_to_u8_vec(ptr: u32, len: u32) -> Vec<u8> {
    return slice::from_raw_parts(ptr as *mut u8, len as usize).to_vec();
}

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
    let vec: Vec<u8> = Vec::<u8>::with_capacity(size);
    let mut v = ManuallyDrop::new(vec);
    return v.as_mut_ptr();
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
