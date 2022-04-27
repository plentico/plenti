import { user, repoUrl } from './auth.js';
import { env } from '../env.js';

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
export async function publish(file, contents, action, encoding) {
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
    const payload = {
        branch: env.cms.branch,
        commit_message: capitalizeFirstLetter(action) + ' ' + file,
        actions: [
            {
                action: action,
                file_path: file,
                encoding: encoding,
                content: contents,
            },
        ],
    };

    const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload),
    });
    if (response.ok) {
        console.log('Successfully published!');
    } else {
        const { error, message } = await response.json();
        throw new Error(`Publish failed: ${error || message}`);
    }
}