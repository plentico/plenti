import { env } from '../../generated/env.js';

export default function evaluateRoute(commitItem) {
    let content_type = window.location.hash.split("/")[1];
    let filename = window.location.hash.split("/")[2];
    let filepath = commitItem.file;
    let fields = JSON.parse(commitItem.contents);
    let default_route = "/" + content_type + "/" + filename;
    let route_pattern = env?.routes?.[content_type] ?? default_route;

    // Replace patterns with actual values
    let route = replaceRouteFields(route_pattern, fields);
    route = route.replace(':filename', filename);
    route = route.replace(':filepath', filepath);
    return route;
}

function replaceRouteFields(route, fields) {
    // Regular expression to match :fields(name) pattern
    const fieldRegex = /:fields\((\w+)\)/g;

    // Replace each match with the corresponding value from the fields object
    return route.replace(fieldRegex, (match, fieldName) => {
        if (fieldName in fields) {
          return slugify(fields[fieldName]);
        }
        // If the field is not found in the object, return the original match
        return match;
    });
}

function slugify(str) {
    str = str.replace(/^\s+|\s+$/g, ''); // trim leading/trailing white space
    str = str.toLowerCase(); // convert string to lowercase
    str = str.replace(/[^a-z0-9 -]/g, '') // remove any non-alphanumeric characters
             .replace(/\s+/g, '-') // replace spaces with hyphens
             .replace(/-+/g, '-'); // remove consecutive hyphens
    return str;
}
