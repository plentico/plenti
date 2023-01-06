const imageExtensions = ['jpg', 'jpeg', 'png', 'webp', 'gif', 'svg', 'avif', 'apng'];
const reImage = new RegExp("^data:image\/(?:" + imageExtensions.join("|") + ")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
export const isImage = file => {
    return imageExtensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reImage.test(file);
}

const docExtensions = ['pdf', 'msword'];
const reDoc = new RegExp("^data:application\/(?:" + docExtensions.join("|") +")(?:;charset=utf-8)?;base64,(?:[A-Za-z0-9]|[+/])+={0,2}");
export const isDoc = file => {
    return docExtensions.includes(file.substr(file.lastIndexOf('.') + 1)) || reDoc.test(file);
}

export const isAsset = file => isImage(file) || isDoc(file);

const reAssetPath = new RegExp("^/?assets/.*\.(" + imageExtensions.join("|") + "|" + docExtensions.join("|") + ")$");
const reImagePath = new RegExp("^/?assets/.*\.(" + imageExtensions.join("|") + ")$");
const reDocPath = new RegExp("^/?assets/.*\.(" + docExtensions.join("|") + ")$");
export const isAssetPath = asset => reAssetPath.test(asset);
export const isImagePath = asset => reImagePath.test(asset);
export const isDocPath = asset => reDocPath.test(asset);