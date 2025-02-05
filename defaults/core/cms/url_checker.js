export const makeUrl = string => {
  try {
    return new URL(string);
  } catch (_) {
    return new URL("https://gitlab.com");
  }
}

export const normalizeRoute = route => {
  // Remove leading and trailing slashes
  return route.replace(/^\/+|\/+$/g, '');
}