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

/**
 * @param {string} file
 * @param {string} content
 */
export async function publish(file, content) {
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
        commit_message: 'Update content',
        actions: [
            {
                action: 'update',
                file_path: file,
                content,
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