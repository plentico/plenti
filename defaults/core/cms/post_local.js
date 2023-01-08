import { env } from '../../generated/env.js';

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