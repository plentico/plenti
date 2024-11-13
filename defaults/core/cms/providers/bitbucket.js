import { env } from '../../../generated/env.js';
import { makeUrl } from '../url_checker.js';
import evaluateRoute from '../route_eval.js';

const repoUrl = makeUrl(env.cms.repo);
const owner = repoUrl.pathname.split('/')[1];
const repo = repoUrl.pathname.split('/')[2];
const apiBaseUrl = `https://api.${repoUrl.host}/2.0/repositories`;
const branch = env.cms.branch;

const capitalizeFirstLetter = string => {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

/**
 * @param {string} file
 * @param {string} contents
 * @param {string} action
 */
export async function commitBitbucket(commitList, shadowContent, action, encoding, user) {
    // Keep track of current user and promise it's availability.
    let currentUser;
    const userAvailable = new Promise(resolve => {
        user.subscribe(user => {
            currentUser = user;
            resolve();
        });
    });

    await userAvailable;
    if (!currentUser.isAuthenticated) {
        throw new Error('Authentication required');
    }

    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${currentUser.tokens.access_token}`,
    };

    for (const commitItem of commitList) {
        //const url = `${apiBaseUrl}/${owner}/${repo}/src/${branch}/` + commitItem.file;
        const url = `${apiBaseUrl}/${owner}/${repo}/src`;

        const makeDataStr = base64Str => base64Str.split(',')[1];
        let message = capitalizeFirstLetter(action) + ' ' + (commitList.length > 1 ? commitList.length + ' files' : commitList[0].file);
        let content = encoding === "base64" ? makeDataStr(commitItem.contents) : btoa(unescape(encodeURIComponent(commitItem.contents)));

        const payload = {
            branch: branch,
            message: message,
            files: commitItem.file,
            content: content,
        };

        let method = action === 'create' ? 'POST' : action === 'update' ? 'PUT' : action === 'delete' ? 'DELETE' : '';

        const response = await fetch(url, {
            method: method,
            headers,
            body: JSON.stringify(payload),
        });
        if (response.ok) {
            if (action === 'create' || action === 'update') {
                shadowContent?.onSave?.();
                // Make sure saving single content file, not list of media items
                if (commitList.length === 1 && commitList[0].file.lastIndexOf('.json') > 0) {
                    let evaluatedRoute = evaluateRoute(commitList[0]);
                    // Redirect only if new route is being created
                    if (evaluatedRoute !== location.pathname) {
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
            const { error, message } = await response.json();
            throw new Error(`Publish failed: ${error || message}`);
        }
    };
}
