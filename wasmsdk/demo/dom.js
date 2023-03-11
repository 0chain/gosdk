const get = id => document.getElementById(id)

const onClick = (id, callback) =>
    document.getElementById(id).addEventListener('click', callback)

const setHtml = (id, html) => document.getElementById(id).innerHTML = html

const readBytes = file => new Promise(resolve => {
  const reader = new FileReader()
  reader.onloadend = () => {
    resolve(new Uint8Array(reader.result))
  }
  reader.readAsArrayBuffer(file)
})