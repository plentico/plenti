import { writable } from 'svelte/store';

/**
 * @param key Item key in snake_case
 * @returns Writable store that is saved to local storage
 */
export function createDataStore(key) {
    key = key.toUpperCase();
    key = "PLENTI_CMS_" + key;
    
    let value = localStorage.getItem(key);
    if (value)
        value = JSON.parse(value);
    else
        value = null;

    const store = writable(value);
    store.subscribe(value => {
        localStorage.setItem(key, JSON.stringify(value));
    });

    return store;
}