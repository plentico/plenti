import { env } from '../../../generated/env.js';
import { commitGitlab } from './gitlab.js';
import { commitGitea } from './gitea.js';
import { postLocal } from './local.js';

// Dispatch a commit to the active CMS provider. Shared by button.svelte and the
// deferred pending-media flow so provider selection lives in exactly one place
// (rather than being duplicated wherever something needs to commit).
export async function commit(commitList, shadowContent, action, encoding, user) {
    const local = env.local ?? false;
    const provider = env.cms.provider.toLowerCase();
    if (local) {
        return postLocal(commitList, shadowContent, action, encoding, user);
    } else if (!provider || provider === "gitlab") {
        return commitGitlab(commitList, shadowContent, action, encoding, user);
    } else if (provider === "gitea" || provider === "forgejo") {
        return commitGitea(commitList, shadowContent, action, encoding, user);
    }
    throw new Error(`Unknown CMS provider: ${provider}`);
}
