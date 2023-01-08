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

export const isMedia = file => isImage(file) || isDoc(file);

const reMediaPath = new RegExp("^/?media/.*\.(" + imageExtensions.join("|") + "|" + docExtensions.join("|") + ")$");
const reImagePath = new RegExp("^/?media/.*\.(" + imageExtensions.join("|") + ")$");
const reDocPath = new RegExp("^/?media/.*\.(" + docExtensions.join("|") + ")$");
export const isMediaPath = media => reMediaPath.test(media);
export const isImagePath = media => reImagePath.test(media);
export const isDocPath = media => reDocPath.test(media);