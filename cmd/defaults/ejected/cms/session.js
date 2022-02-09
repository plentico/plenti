import { writable } from 'svelte/store';

/**
 * @param key Item key in snake_case
 * @returns Writable store that is saved to session storage
 */
export function createSessionStore(key) {
    key = key.toUpperCase();
    key = "PLENTI_CMS_" + key;
    
    let value = sessionStorage.getItem(key);
    if (value)
        value = JSON.parse(value);
    else
        value = null;

    const store = writable(value);
    store.subscribe(value => {
        sessionStorage.setItem(key, JSON.stringify(value));
    });

    return store;
}