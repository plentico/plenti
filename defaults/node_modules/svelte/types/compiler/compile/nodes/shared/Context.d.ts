import { Node, Identifier } from 'estree';
export interface Context {
    key: Identifier;
    name?: string;
    modifier: (node: Node) => Node;
}
export declare function unpack_destructuring(contexts: Context[], node: Node, modifier: (node: Node) => Node): void;
