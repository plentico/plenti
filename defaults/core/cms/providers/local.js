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
        // Per-item action/encoding (falling back to the call-level values) so one
        // commit can mix content (update/text) and media derivatives (create/base64).
        const itemAction = commitItem.action ?? action;
        const itemEncoding = commitItem.encoding ?? encoding;
        // The local /postlocal write overwrites in place, so the provider-neutral
        // 'upsert' (create-or-replace a media derivative) maps to 'create' here —
        // the server validator only knows create/update/delete.
        const wireAction = itemAction === 'upsert' ? 'create' : itemAction;
        body.push({
            action: wireAction,
            encoding: itemEncoding,
            file: commitItem.file,
            contents: itemEncoding === "base64" ? makeDataStr(commitItem.contents) : commitItem.contents
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
        // The local /postlocal endpoint returns plain-text errors (http.Error),
        // not JSON — read text so the message survives instead of throwing on a
        // JSON parse ("Unexpected token ...").
        const message = await response.text();
        throw new Error(`Save failed (${response.status}): ${message.trim() || response.statusText}`);
    }
}
