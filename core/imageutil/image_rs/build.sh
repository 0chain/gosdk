#!/bin/sh

$HOME/.cargo/bin/cargo wasi build --release
cp target/wasm32-wasi/release/image_rs.wasm .

$HOME/.cargo/bin/wasm-pack build --target web

echo "Done!"