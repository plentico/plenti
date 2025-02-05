import { env } from '../../../generated/env.js';
import { normalizeRoute } from '../url_checker.js';
import evaluateRoute from '../route_eval.js';

export async function postLocal(commitList, shadowContent, action, encoding) {
    let url = '/postlocal';
    const headers = {
        'Content-Type': 'application/json; charset=utf-8'
    };
    const makeDataStr = base64Str => base64Str.split(',')[1];
    let body = [];
    commitList.forEach(commitItem => {
        body.push({
            action, action,
            encoding: encoding,
            file: commitItem.file,
            contents: encoding === "base64" ? makeDataStr(commitItem.contents) : commitItem.contents
        });
    });
    const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
    });
    if (response.ok) {
        if (action === 'create' || action === 'update') {
            shadowContent?.onSave?.();
            // Make sure saving single content file, not list of media items
            if (commitList.length === 1 && commitList[0].file.lastIndexOf('.json') > 0) {
                let evaluatedRoute = evaluateRoute(commitList[0]);
                // Redirect only if new route is being created
                if (normalizeRoute(evaluatedRoute) !== normalizeRoute(location.pathname)) {
                    history.pushState({
                        isNew: true,
                        route: evaluatedRoute
                    }, '', evaluatedRoute);
                }
            }
        }
        if (action === 'delete') {
            shadowContent?.onDelete?.();
            history.pushState(null, '', env.baseurl && !env.local ? env.baseurl : '/');
        }
    } else {
        const { error, message } = await response.json();
        throw new Error(`Publish failed: ${error || message}`);
    }
}
