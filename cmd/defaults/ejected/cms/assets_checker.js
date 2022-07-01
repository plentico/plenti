let imageExtensions = ['jpg', 'jpeg', 'png', 'webp', 'gif', 'svg', 'avif', 'apng'];
let reImage = new RegExp("^data:image\/(?:" + imageExtensions.join("|") + ")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
export const isImage = file => {
    return imageExtensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reImage.test(file);
}

let docExtensions = ['pdf', 'msword'];
let reDoc = new RegExp("^data:application\/(?:" + docExtensions.join("|") +")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
export const isDoc = file => {
    return docExtensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reDoc.test(file);
}

export const reAsset = new RegExp("^/?assets/.*\.$");
export const isAsset = asset => reAsset.test(asset);