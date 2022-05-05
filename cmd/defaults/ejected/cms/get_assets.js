import { env } from '../env.js';

/**
 * Base URL without trailing slash
 */
export const baseUrl = env.baseurl.replace(/\/$/, '');

/**
 * Asset list URL
 */
const indexUrl = baseUrl + '/spa/assets/index.json';

export const getAssets = () => {
  return fetch(indexUrl)
    .then(response => response.json());
}