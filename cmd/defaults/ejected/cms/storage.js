class DataStore {
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
        const value = localStorage.getItem(key);
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
        localStorage.setItem(key, JSON.stringify(value));
    }
};

export const storage = new DataStore();