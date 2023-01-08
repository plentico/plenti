import { user, repoUrl } from './auth.js';
import { env } from '../../generated/env.js';

const apiBaseUrl = `${repoUrl.origin}/api/v4`;

// Keep track of current user and promise it's availability.
let currentUser;
const userAvailable = new Promise(resolve => {
    user.subscribe(user => {
        currentUser = user;
        resolve();
    });
});

const capitalizeFirstLetter = string => {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

/**
 * @param {string} file
 * @param {string} contents
 * @param {string} action
 */
export async function publish(commitList, shadowContent, action, encoding) {
    await userAvailable;
    if (!currentUser.isAuthenticated) {
        throw new Error('Authentication required');
    }

    const id = repoUrl.pathname.slice(1);
    const url = `${apiBaseUrl}` +
        `/projects/${encodeURIComponent(id)}` +
        '/repository/commits';
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${currentUser.tokens.access_token}`,
    };

    const makeDataStr = base64Str => base64Str.split(',')[1];

    let actions = [];
    commitList.forEach(commitItem => {
        actions.push({
            action: action,
            file_path: commitItem.file,
            encoding: encoding,
            content: encoding === "base64" ? makeDataStr(commitItem.contents) : commitItem.contents,
        });
    });

    let message = capitalizeFirstLetter(action) + ' ' + (commitList.length > 1 ? commitList.length + ' files' : commitList[0].file);

    const payload = {
        branch: env.cms.branch,
        commit_message: message,
        actions: actions,
    };

    const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload),
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