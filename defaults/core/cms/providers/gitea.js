import { env } from '../../../generated/env.js';
import { makeUrl, normalizeRoute } from '../url_checker.js';
import evaluateRoute from '../route_eval.js';

const repoUrl = makeUrl(env.cms.repo);
const owner = repoUrl.pathname.split('/')[1];
const repo = repoUrl.pathname.split('/')[2];
// Optional override (env.cms.apiBaseUrl). Empty falls back to the original derivation
// (repoUrl.origin + /api/v1), so existing sites are unchanged; set it to route the
// repository API through a same-origin proxy.
const apiBaseUrl = env.cms.apiBaseUrl || `${repoUrl.origin}/api/v1`;

const capitalizeFirstLetter = string => {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

/**
 * @param {string} file
 * @param {string} contents
 * @param {string} action
 */
export async function commitGitea(commitList, shadowContent, action, encoding, user) {
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

    // Set default Gitea User
    let giteaUser = {
        email: 'cms@plenti.co',
        name: 'CMS'
    };

    await fetch(`${apiBaseUrl}` + `/user`, {
        method: 'GET',
        headers
    }).then(response => {
        return response.json();
    }).then(data => {
        // Get actual Gitea User
        giteaUser = data;
    });

    // Gitea's contents API is per-file (no atomic multi-file commit here — that is
    // a separate provider-wide upgrade, see ADR 0001), so a mixed save is N
    // sequential commits. Order by repository path, not action: deferred raw/as-is
    // uploads use 'create', while derivatives use 'upsert'. All media writes must
    // succeed before referencing content is written. A later failure can still
    // leave media behind; this ordering does not make the batch atomic.
    const isMediaWrite = item => (item.action ?? action) !== 'delete'
        && item.file.replace(/^\/+/, '').startsWith('media/');
    const orderedList = [
        ...commitList.filter(isMediaWrite),
        ...commitList.filter(i => !isMediaWrite(i)),
    ];

    for (const commitItem of orderedList) {
        // Per-item action/encoding (falling back to the call-level values) so a
        // single save can mix content (update/text) and media (upsert/base64).
        let itemAction = commitItem.action ?? action;
        const itemEncoding = commitItem.encoding ?? encoding;
        const url = `${apiBaseUrl}/repos/${owner}/${repo}/contents/` + commitItem.file;

        const makeDataStr = base64Str => base64Str.split(',')[1];

        // Resolve the provider-neutral 'upsert' (a derivative that may already exist
        // from a prior session) into create/update via existence: GET the file →
        // 200 = update (use the returned sha), 404 = create. Any other status is a
        // real error (auth/permission/server) and aborts the save — never treated as
        // "absent". A fixed 'create' would 422 on a re-derived path cross-session.
        let resolvedSha;
        if (itemAction === 'upsert') {
            const probe = await fetch(url, { method: 'GET', headers });
            if (probe.ok) {
                const data = await probe.json();
                itemAction = 'update';
                resolvedSha = data.sha;
            } else if (probe.status === 404) {
                itemAction = 'create';
            } else {
                throw new Error(`Publish failed: could not resolve ${commitItem.file} (HTTP ${probe.status})`);
            }
        }

        let message = capitalizeFirstLetter(itemAction) + ' ' + (commitList.length > 1 ? commitList.length + ' files' : commitList[0].file);
        let content = itemEncoding === "base64" ? makeDataStr(commitItem.contents) : btoa(unescape(encodeURIComponent(commitItem.contents)));

        const payload = {
            author: {
                email: giteaUser?.email,
                name: giteaUser.login
            },
            branch: env.cms.branch,
            message: message,
            content: content,
        };

        if (resolvedSha) {
            payload.sha = resolvedSha;
        } else if (itemAction === 'update' || itemAction === 'delete') {
            // Get details about existing file from Gitea
            await fetch(url, {
                method: 'GET',
                headers,
            }).then(response => {
                return response.json();
            }).then(data => {
                // Set the required SHA in payload
                payload.sha = data.sha;
            });
        }

        let method = itemAction === 'create' ? 'POST' : itemAction === 'update' ? 'PUT' : itemAction === 'delete' ? 'DELETE' : '';

        const response = await fetch(url, {
            method: method,
            headers,
            body: JSON.stringify(payload),
        });
        // Per-file failure aborts the whole save BEFORE any success signal. With
        // media ordered first, this never leaves content pointing at a missing
        // derivative, and the UI is never told "saved" for a partial write.
        if (!response.ok) {
            const { error, message } = await response.json();
            throw new Error(`Publish failed: ${error || message}`);
        }
    }

    // Commit-level success only — fire AFTER every file in the sequential save has
    // committed. Calling onSave per item would signal success after the media write
    // but before a failing content write, contradicting the retryable guarantee.
    if (action === 'create' || action === 'update') {
        shadowContent?.onSave?.();
        // Make sure saving single content file, not list of media items
        if (commitList.length === 1 && commitList[0].file.lastIndexOf('.json') > 0) {
            let evaluatedRoute = evaluateRoute(commitList[0]);
            // Redirect only if new route is being created
            if (normalizeRoute(evaluatedRoute) !== normalizeRoute(location.pathname)) {
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
}
