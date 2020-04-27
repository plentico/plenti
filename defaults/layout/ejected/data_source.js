import nodes from './nodes.js';

class DataSource {

  constructor() {}

  static getNode(uri) {
    return nodes.find(node => node.path == uri);
  }

  static getAllNodes() {
    return nodes;
  }
}

export default DataSource;
