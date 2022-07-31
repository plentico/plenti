export async function postLocal(commitList, action, encoding) {
    console.log("post_local.js fired");
    let url = 'http://localhost:3000/postlocal';
    const headers = {
        'Content-Type': 'application/json; charset=utf-8'
    };
    let body = [];
    commitList.forEach(item => {
        body.push({
            file: item.file,
            contents: item.contents
        });
    });
    const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
        //body: body,
    });
    if (response.ok) {
        console.log("Saved locally!");
    } else {
        const { error, message } = await response.json();
        throw new Error(`Publish failed: ${error || message}`);
    }
}