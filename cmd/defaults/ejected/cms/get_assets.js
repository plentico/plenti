export const getAssets = dir => {
  return fetch(dir)
    .then(response => response.text())
    .then(data => {
        let parser = new DOMParser();
        let doc = parser.parseFromString(data, 'text/html');
        return doc.querySelectorAll('a');
    });
}