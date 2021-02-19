import { DecodedSourceMap, RawSourceMap } from '@ampproject/remapping/dist/types/types';
import { SourceMap } from 'magic-string';
declare type SourceLocation = {
    line: number;
    column: number;
};
export declare function sourcemap_add_offset(map: DecodedSourceMap, offset: SourceLocation, source_index: number): void;
export declare class StringWithSourcemap {
    string: string;
    map: DecodedSourceMap;
    constructor(string?: string, map?: DecodedSourceMap);
    /**
     * concat in-place (mutable), return this (chainable)
     * will also mutate the `other` object
     */
    concat(other: StringWithSourcemap): StringWithSourcemap;
    static from_processed(string: string, map?: DecodedSourceMap): StringWithSourcemap;
    static from_source(source_file: string, source: string, offset?: SourceLocation): StringWithSourcemap;
}
export declare function combine_sourcemaps(filename: string, sourcemap_list: Array<DecodedSourceMap | RawSourceMap>): RawSourceMap;
export declare function apply_preprocessor_sourcemap(filename: string, svelte_map: SourceMap, preprocessor_map_input: string | DecodedSourceMap | RawSourceMap): SourceMap;
export {};
