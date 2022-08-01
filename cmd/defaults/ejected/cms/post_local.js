export async function postLocal(commitList, action, encoding) {
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
        console.log("Saved locally!");
    } else {
        const { error, message } = await response.json();
        throw new Error(`Publish failed: ${error || message}`);
    }
}