export const getLinks = assetsDir => {
  return fetch(assetsDir)
    .then(response => response.text())
    .then(data => {
        let parser = new DOMParser();
        let doc = parser.parseFromString(data, 'text/html');
        return doc.querySelectorAll('a');
    });
}