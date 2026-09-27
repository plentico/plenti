import { env } from '../../../generated/env.js';
import { makeUrl, normalizeRoute } from '../url_checker.js';
import evaluateRoute from '../route_eval.js';

const repoUrl = makeUrl(env.cms.repo);
// Optional override (env.cms.apiBaseUrl). Empty falls back to the original derivation
// (repoUrl.origin + /api/v4), so existing sites are unchanged; set it to route the
// repository API through a same-origin proxy.
const apiBaseUrl = env.cms.apiBaseUrl || `${repoUrl.origin}/api/v4`;

const capitalizeFirstLetter = string => {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

/**
 * @param {string} file
 * @param {string} contents
 * @param {string} action
 */
export async function commitGitlab(commitList, shadowContent, action, encoding, user) {
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

    const id = repoUrl.pathname.slice(1);
    const projectUrl = `${apiBaseUrl}/projects/${encodeURIComponent(id)}`;
    const url = `${projectUrl}/repository/commits`;
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${currentUser.tokens.access_token}`,
    };

    const makeDataStr = base64Str => base64Str.split(',')[1];
    const branch = env.cms.branch;

    // Resolve the provider-neutral 'upsert' (a media derivative that may already
    // exist from a prior session) into GitLab's explicit create/update BEFORE the
    // atomic commit: HEAD the file → 404 = create, 200 = update (+ last_commit_id,
    // required by the batch Commit API for an update). Any other status is a real
    // error (auth/permission/server) and aborts the whole save — never silently
    // treated as "absent". All resolutions happen up front so the commit stays one
    // atomic request (all-or-nothing), preserving GitLab's strongest property.
    const resolveUpsert = async file => {
        const head = await fetch(
            `${projectUrl}/repository/files/${encodeURIComponent(file)}?ref=${encodeURIComponent(branch)}`,
            { method: 'HEAD', headers },
        );
        if (head.status === 404) return { action: 'create' };
        if (head.ok) return { action: 'update', last_commit_id: head.headers.get('X-Gitlab-Last-Commit-Id') };
        throw new Error(`Publish failed: could not resolve ${file} (HTTP ${head.status})`);
    };

    let actions = [];
    for (const commitItem of commitList) {
        // Per-item action/encoding (falling back to the call-level values) so one
        // atomic commit can mix content (update/text) and media (upsert/base64).
        const itemAction = commitItem.action ?? action;
        const itemEncoding = commitItem.encoding ?? encoding;
        const entry = {
            action: itemAction,
            file_path: commitItem.file,
            encoding: itemEncoding,
            content: itemEncoding === "base64" ? makeDataStr(commitItem.contents) : commitItem.contents,
        };
        if (itemAction === 'upsert') {
            const resolved = await resolveUpsert(commitItem.file);
            entry.action = resolved.action;
            if (resolved.last_commit_id) entry.last_commit_id = resolved.last_commit_id;
        }
        actions.push(entry);
    }

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
    } else {
        const { error, message } = await response.json();
        throw new Error(`Publish failed: ${error || message}`);
    }
}
