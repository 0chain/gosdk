#!/bin/sh

cargo wasi build --release
cp target/wasm32-wasi/release/image_rs.wasm .

echo "Done!"