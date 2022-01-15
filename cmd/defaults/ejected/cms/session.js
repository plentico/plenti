class SessionStore {
    prefixKey(key) {
        key = key.toUpperCase();
        key = "PLENTI_CMS_" + key;
        return key;
    }

    /**
     * @param {string} key
     */
    get(key) {
        key = this.prefixKey(key);
        const value = sessionStorage.getItem(key);
        if (value)
            return JSON.parse(value);
        else
            return null;
    }

    /**
     * @param {string} key
     * @param {any} value
     */
    set(key, value) {
        key = this.prefixKey(key);
        sessionStorage.setItem(key, JSON.stringify(value));
    }
};

export const session = new SessionStore();