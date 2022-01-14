<script context="module">
    import { session } from './session.svelte';
    import { storage } from './storage.svelte';

    const settings = {
        "server": "gitlab.com",
        "redirectUrl": "http://localhost:3000/admin/",
        "appId": "6307e5b50e8631f9dc67ed2e982dc5f02159d84a4b7cbd054a9f76eb109e5990"
    };

    const generateString = () => {
        const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.';
        const randomValues = Array.from(crypto.getRandomValues(new Uint8Array(128)));
        return randomValues.map(val => chars[val % chars.length]).join('');
    }

    const hash = async text => {
        const encoder = new TextEncoder();
        const data = encoder.encode(text);
        const digest = await crypto.subtle.digest('SHA-256', data);
        const binary = String.fromCharCode(...new Uint8Array(digest));
        return btoa(binary)
            .split('=')[0]
            .replace(/\\+/g, '-')
            .replace(/\\//g, '_');
    }

    export const requestAuthCode = async () => {
        const state = generateString();
        session.set('gitlab_state', state);

        const codeVerifier = generateString();
        session.set('gitlab_code_verifier', codeVerifier);
        const codeChallenge = await hash(codeVerifier);

        const { server, redirectUrl, appId } = settings;
        window.location.href = "https://" + server + "/oauth/authorize"
            + "?client_id=" + encodeURIComponent(appId) 
            + "&redirect_uri=" + encodeURIComponent(redirectUrl)
            + "&response_type=code"
            + "&state=" + encodeURIComponent(state) 
            + "&code_challenge=" + encodeURIComponent(codeChallenge) 
            + "&code_challenge_method=S256";    
    }

    export const requestAccessToken = async code => {
        const { server, redirectUrl, appId } = settings;
        const codeVerifier = session.get('gitlab_code_verifier');
        const response = await fetch("https://" + server + "/oauth/token"
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
        storage.set('gitlab_tokens', tokens);
        location.href = location.pathname;
    }

    export const requestRefreshToken = async () => {
        const codeVerifier = session.get('code_verifier');
        if (!codeVerifier) {
            throw new Error("Code verifier not saved to session storage");
        }
        const response = await fetch("https://" + server + "/oauth/token"
            + "?client_id=" + encodeURIComponent(appId)
            + "&refresh_token=" + encodeURIComponent(refreshToken)
            + "&grant_type=refresh_token"
            + "&redirect_uri=" + encodeURIComponent(redirectUrl)
            + "&code_verifier=" + encodeURIComponent(codeVerifier),
            { method: 'POST' }
        );
        const tokens = await response.json();
        if (tokens.error) {
            throw new Error(tokens.error_description);
        }
        storage.set('gitlab_tokens', tokens);
    }

</script>