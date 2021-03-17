declare module 'navaid' {
	type Promisable<T> = T | Promise<T>;

	export type Params = Record<string, string>;
	export type UnknownHandler = (uri: string) => void;
	export type RouteHandler<T=Params> = (params?: T) => Promisable<any>;

	export interface Router {
		format(uri: string): string | false;
		route(uri: string, replace?: boolean): void;
		on<T=Params>(pattern: string, handler: RouteHandler<T>): Router;
		run(uri?: string): Router;
		listen(uri?: string): Router;
		unlisten?: VoidFunction;
	}

	export default function (base?: string, on404?: UnknownHandler): Router;
}
