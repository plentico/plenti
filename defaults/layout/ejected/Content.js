import nodes from './nodes.js';

  class Content {
  
    constructor() {}
  
    static getNode(uri) {
      let content;
      nodes.map(node => {
        if (node.path == uri) {
          content = node;
        }
      });
      return content ? content : '';
    }
  
    static getAllNodes() {
      let content = nodes.map(node => {
        return node;
      });
      return content;
    }
  }
  
  export default Content;