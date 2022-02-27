import { readable } from 'svelte/store';
import { createSessionStore } from './session.svelte';
import { createDataStore } from './storage.svelte';
import { env } from '../env.js';

const repoUrl = new URL(env.cms.repo);

const settings = {
    server: repoUrl.origin,
    group: repoUrl.pathname.split('/')[1],
    repository: repoUrl.pathname.split('/')[2],
    redirectUrl: env.cms.redirectUrl,
    appId: env.cms.appId
};

const tokenStore = createDataStore('gitlab_tokens');
let tokens, isExpired;
tokenStore.subscribe(value => {
    tokens = value;
    isExpired = tokens && Date.now() > (tokens.created_at + tokens.expires_in) * 1000;
});

const codeVerifierStore = createDataStore('gitlab_code_verifier');
let codeVerifier;
codeVerifierStore.subscribe(value => codeVerifier = value);

const stateStore = createSessionStore('gitlab_state');
let state;
stateStore.subscribe(value => state = value);

const getUser = () => ({
    isBeingAuthenticated: Boolean(state) || (tokens && isExpired),
    isAuthenticated: tokens && !isExpired,
    tokens,

    finishAuthentication(params) {
        if (params && state && params.get('state') === state) {
            stateStore.set(null);
            history.replaceState(null, '', location.pathname);
            return requestAccessToken(params.get('code'));
        }

        if (tokens && isExpired) {
            return requestRefreshToken();
        }

        console.error('Invalid parameters or state');
    },

    login() {
        return requestAuthCode();
    },

    logout() {
        tokenStore.set(null);
        codeVerifierStore.set(null);
    },
});
export const user = readable(getUser(), set => {
    tokenStore.subscribe(() => set(getUser()));
    stateStore.subscribe(() => set(getUser()));
});

const generateString = () => {
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.';
    const randomValues = Array.from(crypto.getRandomValues(new Uint8Array(128)));
    return randomValues.map(val => chars[val % chars.length]).join('');
};

const hash = async text => {
    const encoder = new TextEncoder();
    const data = encoder.encode(text);
    const digest = await crypto.subtle.digest('SHA-256', data);
    const binary = String.fromCharCode(...new Uint8Array(digest));
    return btoa(binary)
        .split('=')[0]
        .replace(/\+/g, '-')
        .replace(/\//g, '_');
};

const requestAuthCode = async () => {
    stateStore.set(generateString());
    codeVerifierStore.set(generateString());
    const codeChallenge = await hash(codeVerifier);

    const { server, redirectUrl, appId } = settings;
    window.location.href = server + "/oauth/authorize"
        + "?client_id=" + encodeURIComponent(appId) 
        + "&redirect_uri=" + encodeURIComponent(redirectUrl)
        + "&response_type=code"
        + "&state=" + encodeURIComponent(state) 
        + "&code_challenge=" + encodeURIComponent(codeChallenge) 
        + "&code_challenge_method=S256";    
};

const requestAccessToken = async code => {
    const { server, redirectUrl, appId } = settings;
    const response = await fetch(server + "/oauth/token"
        + "?client_id=" + encodeURIComponent(appId) 
        + "&code=" + encodeURIComponent(code) 
        + "&grant_type=authorization_code"
        + "&redirect_uri=" + encodeURIComponent(redirectUrl)
        + "&code_verifier=" + encodeURIComponent(codeVerifier),
        { method: 'POST' }
    );
    const tokens = await response.json();
    if (tokens.error) {
        throw new Error(tokens.error_description);
    }
    tokenStore.set(tokens);
};

const requestRefreshToken = async () => {
    const { server, redirectUrl, appId } = settings;
    if (!codeVerifier) {
        throw new Error("Code verifier not saved to session storage");
    }
    const response = await fetch(server + "/oauth/token"
        + "?client_id=" + encodeURIComponent(appId)
        + "&refresh_token=" + encodeURIComponent(tokens.refresh_token)
        + "&grant_type=refresh_token"
        + "&redirect_uri=" + encodeURIComponent(redirectUrl)
        + "&code_verifier=" + encodeURIComponent(codeVerifier),
        { method: 'POST' }
    );
    const refreshedTokens = await response.json();
    if (refreshedTokens.error) {
        throw new Error(refreshedTokens.error_description);
    }
    tokenStore.set(refreshedTokens);
};