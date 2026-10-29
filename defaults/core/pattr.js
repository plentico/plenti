/*!
 * Pattr v1.0.0
 * A lightweight reactive framework
 * https://github.com/plentico/pattr
 * MIT License
 */
window.Pattr = {
    _templateScopeCounter: 0,

    /**
     * Creates a shadow object for a getter/setter object assigned via a non-sync p-scope.
     *
     * Getter/setter objects produced by pico (e.g. `content = { get name() {...}, set name(v) {...} }`)
     * have setters that are closure-bound to the PARENT scope's proxy variables. When assigned to a
     * child component's regular p-scope (not :sync), any write to content[key] would fire the original
     * setter which reaches back into the parent's scope, bypassing pattr's proxy entirely.
     *
     * The shadow object:
     *   - Reads through to the original getter (live: reflects parent changes until overridden locally)
     *   - Intercepts setters so writes stay in a local `localOverrides` map and never touch the parent
     *
     * Returns null if the object has no getter-defined properties (not a getter/setter proxy).
     */
    /**
     * Creates (or re-targets) a shadow for a getter/setter proxy object.
     *
     * existingLocalOverrides — pass the `_p_localOverrides` from a previously-created
     * shadow for the same variable so that child-local writes survive a parent re-render
     * that produces a new getter/setter object with a different reference.
     */
    _createGetterSetterShadow(obj, existingLocalOverrides = null, existingExplicitKeys = null) {
        if (obj === null || typeof obj !== 'object') return null;
        const descs = Object.getOwnPropertyDescriptors(obj);
        const hasGetters = Object.values(descs).some(d => typeof d.get === 'function');
        if (!hasGetters) return null;

        // Build localOverrides from ONLY the explicitly-written child overrides (not
        // pre-copies from a previous shadow). Pre-copies are refreshed from the new obj
        // so that parent mutations to arrays propagate to the child unless the child has
        // explicitly written its own value for that key.
        const explicitKeys = new Set(existingExplicitKeys || []);
        const localOverrides = {};
        if (existingLocalOverrides) {
            for (const key of explicitKeys) {
                if (key in existingLocalOverrides) {
                    localOverrides[key] = existingLocalOverrides[key];
                }
            }
        }

        // Track the parent array reference we last pre-copied from, keyed by property name.
        // This lets the getter detect when the parent has replaced an array (new reference)
        // vs an in-place mutation (same reference), and refresh the copy accordingly.
        const preCopyRefs = {};

        const shadow = {};
        for (const [propKey, desc] of Object.entries(descs)) {
            if (typeof desc.get === 'function') {
                const getter = desc.get;
                Object.defineProperty(shadow, propKey, {
                    get: () => {
                        // Explicit child write always wins.
                        if (explicitKeys.has(propKey)) {
                            return localOverrides[propKey];
                        }
                        const parentVal = getter.call(obj);
                        // For arrays: keep a local copy so in-place child mutations
                        // (e.g. arr[0] = x) stay isolated from the parent's actual array.
                        // Invalidate the copy whenever the parent provides a NEW reference
                        // (e.g. cats = [...cats, 'NewCat']) so the child always sees the
                        // current parent value without needing the shadow to be recreated.
                        if (Array.isArray(parentVal)) {
                            if (preCopyRefs[propKey] !== parentVal) {
                                localOverrides[propKey] = [...parentVal];
                                preCopyRefs[propKey] = parentVal;
                            }
                            return localOverrides[propKey];
                        }
                        // Scalars and non-array objects read through live.
                        return parentVal;
                    },
                    // Writes go to localOverrides only — never call the original setter.
                    // Clear the pre-copy reference so a future read doesn't overwrite the
                    // explicit child value with a stale copy.
                    set: (v) => { localOverrides[propKey] = v; explicitKeys.add(propKey); delete preCopyRefs[propKey]; },
                    enumerable: true,
                    configurable: true,
                });
            } else if ('value' in desc) {
                shadow[propKey] = desc.value;
            }
        }

        // Expose metadata so callers can preserve explicit child writes across re-shadowing.
        Object.defineProperty(shadow, '_p_isShadow', { value: true, enumerable: false, configurable: false });
        Object.defineProperty(shadow, '_p_localOverrides', { value: localOverrides, enumerable: false, configurable: false });
        Object.defineProperty(shadow, '_p_explicitOverrideKeys', { value: explicitKeys, enumerable: false, configurable: false });

        return shadow;
    },

    directives: {
        'p-text': (el, value, modifiers = {}) => {
            let text = String(value);
            
            // Apply trim modifier
            if (modifiers.trim && modifiers.trim.length > 0) {
                const maxLength = parseInt(modifiers.trim[0]) || 100;
                if (text.length > maxLength) {
                    text = text.substring(0, maxLength) + '...';
                }
            }
            
            el.innerText = text;
        },
        'p-html': (el, value, modifiers = {}) => {
            let html = value;
            
            // Apply allow filter first (if present)
            if (modifiers.allow && modifiers.allow.length > 0) {
                const allowedTags = modifiers.allow;
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = html;
                
                // Recursively filter elements
                const filterNode = (node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        const tagName = node.tagName.toLowerCase();
                        if (!allowedTags.includes(tagName)) {
                            // Replace with filtered children (unwrap the tag but keep allowed nested tags)
                            const fragment = document.createDocumentFragment();
                            Array.from(node.childNodes).forEach(child => {
                                const filteredChild = filterNode(child);
                                if (filteredChild) fragment.appendChild(filteredChild);
                            });
                            return fragment;
                        }
                        // Keep element but filter children
                        const filtered = node.cloneNode(false);
                        Array.from(node.childNodes).forEach(child => {
                            const filteredChild = filterNode(child);
                            if (filteredChild) filtered.appendChild(filteredChild);
                        });
                        return filtered;
                    }
                    return node.cloneNode();
                };
                
                const filtered = document.createElement('div');
                Array.from(tempDiv.childNodes).forEach(child => {
                    const filteredChild = filterNode(child);
                    if (filteredChild) filtered.appendChild(filteredChild);
                });
                html = filtered.innerHTML;
            }
            
            // Apply trim modifier (counts only text, preserves HTML tags)
            if (modifiers.trim && modifiers.trim.length > 0) {
                const maxLength = parseInt(modifiers.trim[0]) || 100;
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = html;
                
                let charCount = 0;
                let truncated = false;
                
                // Recursively traverse and trim while preserving HTML structure
                const trimNode = (node) => {
                    if (truncated) return null;
                    
                    if (node.nodeType === Node.TEXT_NODE) {
                        const text = node.textContent;
                        const remaining = maxLength - charCount;
                        
                        if (text.length <= remaining) {
                            charCount += text.length;
                            return node.cloneNode();
                        } else {
                            // This text node exceeds limit
                            truncated = true;
                            const trimmedText = text.substring(0, remaining) + '...';
                            return document.createTextNode(trimmedText);
                        }
                    } else if (node.nodeType === Node.ELEMENT_NODE) {
                        const cloned = node.cloneNode(false);
                        for (let child of node.childNodes) {
                            const trimmedChild = trimNode(child);
                            if (trimmedChild) {
                                cloned.appendChild(trimmedChild);
                            }
                            if (truncated) break;
                        }
                        return cloned;
                    }
                    return node.cloneNode();
                };
                
                const result = document.createElement('div');
                for (let child of tempDiv.childNodes) {
                    const trimmedChild = trimNode(child);
                    if (trimmedChild) {
                        result.appendChild(trimmedChild);
                    }
                    if (truncated) break;
                }
                
                html = result.innerHTML;
            }
            
            el.innerHTML = html;
        },
        'p-show': (el, value) => {
            el.style.display = value ? '' : 'none'
        },
        'p-style': (el, value) => {
            if (typeof value === 'string') {
                el.style.cssText = value;
            } else if (typeof value === 'object' && value !== null) {
                Object.assign(el.style, value);
            }
        },
        'p-class': (el, value, modifiers = {}) => {
            // Check for .replace modifier - completely replaces all classes
            const shouldReplace = modifiers.replace !== undefined;

            if (shouldReplace) {
                // Replace mode: p-class.replace completely controls the className
                if (typeof value === 'string') {
                    el.className = value;
                } else if (Array.isArray(value)) {
                    el.className = value.join(' ');
                } else if (typeof value === 'object' && value !== null) {
                    el.className = Object.keys(value)
                        .filter(key => value[key])
                        .join(' ');
                }
                return;
            }

            // Merge mode: preserve static classes, add/remove dynamic ones
            
            // Calculate new dynamic classes from the value FIRST
            let newDynamicClasses = new Set();
            // Track ALL potential dynamic class names (for collision detection with object format)
            let allPotentialDynamicClasses = new Set();
            
            if (typeof value === 'string' && value) {
                value.split(' ').forEach(c => {
                    if (c) {
                        newDynamicClasses.add(c);
                        allPotentialDynamicClasses.add(c);
                    }
                });
            } else if (Array.isArray(value)) {
                value.forEach(c => {
                    if (c) {
                        newDynamicClasses.add(c);
                        allPotentialDynamicClasses.add(c);
                    }
                });
            } else if (typeof value === 'object' && value !== null) {
                Object.keys(value).forEach(key => {
                    // All keys are potential dynamic classes (for collision detection)
                    allPotentialDynamicClasses.add(key);
                    // Only add to active dynamic classes if value is truthy
                    if (value[key]) newDynamicClasses.add(key);
                });
            }

            // Initialize static classes tracking on first run
            // IMPORTANT: Exclude any classes that are potential dynamic classes (collisions)
            if (!el._p_staticClasses) {
                const currentClasses = el.className ? el.className.split(' ').filter(c => c) : [];
                // Exclude ALL potential dynamic class names (not just currently active ones)
                // This ensures that for object format like {collapsed: false}, 'collapsed'
                // is still excluded from static classes even though it's not currently active
                el._p_staticClasses = currentClasses.filter(c => !allPotentialDynamicClasses.has(c));
            }
            if (!el._p_dynamicClasses) {
                el._p_dynamicClasses = new Set();
            }

            // Preserve classes added by other scripts while removing stale dynamic classes.
            const currentClasses = el.className ? el.className.split(' ').filter(c => c) : [];
            const externalClasses = currentClasses.filter(c =>
                !el._p_staticClasses.includes(c) &&
                !el._p_dynamicClasses.has(c) &&
                !allPotentialDynamicClasses.has(c)
            );

            // Update tracked dynamic classes
            el._p_dynamicClasses = newDynamicClasses;

            // Build final class list:
            // 1. Start with static classes (already excludes collisions from first run)
            // 2. Add all current dynamic classes
            // 3. Keep classes managed outside of p-class
            const finalClasses = [...el._p_staticClasses, ...el._p_dynamicClasses, ...externalClasses];

            el.className = finalClasses.join(' ');
        },
        'p-model': (el, value) => {
            if (el.type === 'checkbox') {
                el.checked = !!value;
            } else if (el.type === 'radio') {
                el.checked = (el.value === value);
            } else {
                // If value is an array, join with commas for display
                if (Array.isArray(value)) {
                    el.value = value.join(', ');
                } else {
                    el.value = value;
                }
            }
        },
        'p-attr': (el, value, modifiers = {}) => {
            // Two usage modes:
            // 1. Single attribute: p-attr:data-id="userId" or p-attr:href="link"
            // 2. Multiple attributes: p-attr="{ 'data-id': userId, 'data-name': userName }"
            
            // Check if this is single attribute mode (has modifier parts)
            const modifierKeys = Object.keys(modifiers);
            if (modifierKeys.length > 0) {
                // Single attribute mode: p-attr:data-id="value"
                // The attribute name is the first modifier
                const attrName = modifierKeys[0];
                if (value === null || value === undefined || value === false) {
                    el.removeAttribute(attrName);
                } else {
                    el.setAttribute(attrName, String(value));
                }
            } else if (typeof value === 'object' && value !== null) {
                // Multiple attributes mode: p-attr="{ 'data-id': userId }"
                Object.keys(value).forEach(attrName => {
                    const attrValue = value[attrName];
                    if (attrValue === null || attrValue === undefined || attrValue === false) {
                        el.removeAttribute(attrName);
                    } else {
                        el.setAttribute(attrName, String(attrValue));
                    }
                });
            }
        },
    },
    
    parseDirectiveModifiers(attrName) {
        // Parse: p-html:trim.300:allow.p.h1.h2
        // Returns: { directive: 'p-html', modifiers: { trim: ['300'], allow: ['p', 'h1', 'h2'] } }
        const parts = attrName.split(':');
        const directive = parts[0];
        const modifiers = {};
        
        // Parse each modifier group
        for (let i = 1; i < parts.length; i++) {
            const modParts = parts[i].split('.');
            const modName = modParts[0];
            const modValues = modParts.slice(1);
            modifiers[modName] = modValues;
        }
        
        return { directive, modifiers };
    },

    // ==================== INITIALIZATION ====================
    
    async start() {
        this.root = document.documentElement;

        // Load root data (props from CMS). Large shared props can be externalized
        // by the SSR renderer so they are downloaded once instead of embedded in
        // every generated page. Fetch them before creating any reactive scopes.
        const rootDataElement = document.getElementById("p-root-data");
        const rootDataJsonString = rootDataElement?.textContent;
        try {
            this.rawData = JSON.parse(rootDataJsonString || '{}');
        } catch (e) {
            console.error("Error parsing root data JSON:", e);
            this.rawData = {};
        }

        let externalProps = {};
        try {
            externalProps = JSON.parse(rootDataElement?.getAttribute('data-external-props') || '{}');
        } catch (e) {
            console.error("Error parsing external root data sources:", e);
        }
        const externalRequests = Object.entries(externalProps);
        // Keep initialization synchronous for sites with no external props.
        // Existing callers are not required to await start() in that case.
        if (externalRequests.length > 0) {
            await Promise.all(externalRequests.map(async ([name, url]) => {
                try {
                    const response = await fetch(url);
                    if (!response.ok) {
                        throw new Error(`${response.status} ${response.statusText}`);
                    }
                    this.rawData[name] = await response.json();
                } catch (e) {
                    console.error(`Error loading external root data prop "${name}" from ${url}:`, e);
                }
            }));
        }

        this.buildScopeData(this.root, this.rawData);
        this.data = this.observe(this.rawData);
        this.walkDom(this.root, this.data, true);
        this.refreshAllLoops();
    },

    generateDomPathId(el) {
        const path = [];
        let node = el;
        while (node && node.tagName !== 'HTML') {
            const parent = node.parentElement;
            if (parent) {
                const index = Array.from(parent.children).indexOf(node);
                path.unshift(`${node.tagName}[${index}]`);
            }
            node = parent;
        }
        return path.join('/');
    },

    buildScopeData(el, parentData) {
        let currentData = parentData;
        // Check for p-scope attribute (with or without modifiers like p-scope:sync)
        const pScopeAttr = Array.from(el.attributes).find(attr => attr.name.startsWith('p-scope'));
        if (pScopeAttr) {
            // Use explicit p-id if provided, otherwise generate from DOM position
            const dataId = el.getAttribute('p-id') || this.generateDomPathId(el);
            if (!parentData._p_children) {
                parentData._p_children = {};
            }
            if (!parentData._p_children[dataId]) {
                parentData._p_children[dataId] = {};
            }
            currentData = parentData._p_children[dataId]; 
            currentData._p_scope = el.getAttribute(pScopeAttr.name);
            // Store the ID for later use
            currentData._p_id = dataId;
        }
        let child = el.firstElementChild;
        while (child) {
            this.buildScopeData(child, currentData); 
            child = child.nextElementSibling;
        }
    },

    // ==================== REACTIVE PROXY ====================

    observe(data, parentScope, syncVars = []) {
        const localTarget = data;
        let proxyTarget = localTarget;
        if (parentScope) {
            proxyTarget = Object.create(parentScope._p_target || parentScope);
            Object.assign(proxyTarget, localTarget);
        }
        
        const syncVarsSet = new Set(syncVars);
        const parentTarget = parentScope?._p_target || parentScope;
        
        // Track which objects have been wrapped to avoid infinite recursion
        const wrappedObjects = new WeakSet();
        
        const createDeepProxy = (obj, isSynced = false, syncKey = null) => {
            if (obj === null || typeof obj !== 'object') {
                return obj;
            }
            
            // Avoid re-wrapping the same object
            if (wrappedObjects.has(obj)) {
                return obj;
            }
            
            // Don't wrap primitive wrappers, dates, regexes, etc.
            if (obj instanceof Date || obj instanceof RegExp || obj instanceof Map || obj instanceof Set) {
                return obj;
            }
            
            // Create wrapper that intercepts get/set on properties
            const wrapper = new Proxy(obj, {
                get: (target, key) => {
                    if (key === '_p_target') {
                        return target;
                    }
                    if (key === '_p_syncParent') {
                        return parentTarget;
                    }
                    if (key === '_p_syncKey') {
                        return syncKey;
                    }
                    
                    const value = target[key];
                    
                    // For synced objects, wrap nested objects too
                    if (isSynced && value !== null && typeof value === 'object') {
                        return createDeepProxy(value, true, syncKey);
                    }
                    
                    return value;
                },
                set: (target, key, value) => {
                    // If setting a nested object, wrap it with a deep proxy
                    if (value !== null && typeof value === 'object') {
                        value = createDeepProxy(value, isSynced, syncKey);
                    }
                    
                    target[key] = value;
                    
                    // If this is a synced object, propagate to parent
                    if (isSynced && parentTarget && syncKey) {
                        const parentVal = parentTarget[syncKey];
                        if (parentVal && typeof parentVal === 'object') {
                            // Also wrap parent's nested object if needed
                            if (value !== null && typeof value === 'object' && !parentVal[key]._p_isProxy) {
                                parentVal[key] = createDeepProxy(value, isSynced, syncKey);
                            } else {
                                parentVal[key] = value;
                            }
                        }
                    }
                    
                    this.walkDom(this.root, this.data, false);
                    return true;
                },
                has: (target, key) => {
                    return key in target;
                }
            });
            
            wrappedObjects.add(obj);
            return wrapper;
        };
        
        // Wrap synced variables with deep proxies
        if (parentTarget) {
            for (const syncVar of syncVarsSet) {
                if (syncVar in proxyTarget) {
                    const value = proxyTarget[syncVar];
                    if (value !== null && typeof value === 'object') {
                        proxyTarget[syncVar] = createDeepProxy(value, true);
                    }
                }
                // Also wrap the parent's synced object
                if (syncVar in parentTarget) {
                    const parentValue = parentTarget[syncVar];
                    if (parentValue !== null && typeof parentValue === 'object') {
                        parentTarget[syncVar] = createDeepProxy(parentValue, true);
                    }
                }
            }
        }
        
        const proxy = new Proxy(proxyTarget, {
            get: (target, key) => {
                if (key === '_p_target') {
                    return target;
                }
                return target[key];
            },
            set: (target, key, value) => {
                // If setting a synced object, wrap it with a deep proxy
                if (syncVarsSet.has(key) && value !== null && typeof value === 'object') {
                    value = createDeepProxy(value, true, key);
                }

                // For non-sync assignments of getter/setter objects, create a shadow so that
                // setter calls from within the child scope don't write through to the parent's
                // closure variables (which would bypass p-scope isolation).
                if (!syncVarsSet.has(key) && value !== null && typeof value === 'object') {
                    const shadow = this._createGetterSetterShadow(value);
                    if (shadow) value = shadow;
                }
                
                target[key] = value;
                
                // If this is a synced variable, also update the parent scope
                if (syncVarsSet.has(key) && parentTarget && key in parentTarget) {
                    parentTarget[key] = value;
                }
                
                this.walkDom(this.root, this.data, false);
                return true;
            },
            has: (target, key) => {
                return key in target;
            }
        });
        return proxy;
    },

    // ==================== SCOPE MANAGEMENT ====================

    /**
     * Splits p-scope expression into statements, respecting braces and parentheses
     */
    splitPScopeStatements(pScopeExpr) {
        const statements = [];
        let current = '';
        let braceDepth = 0;
        let parenDepth = 0;
        let inString = false;
        let stringChar = null;
        
        for (let i = 0; i < pScopeExpr.length; i++) {
            const char = pScopeExpr[i];
            const prevChar = i > 0 ? pScopeExpr[i - 1] : null;
            
            // Handle string boundaries
            if ((char === '"' || char === "'" || char === '`') && prevChar !== '\\') {
                if (!inString) {
                    inString = true;
                    stringChar = char;
                } else if (stringChar === char) {
                    inString = false;
                    stringChar = null;
                }
            }
            
            // Track brace/paren depth (only when not in string)
            if (!inString) {
                if (char === '{') braceDepth++;
                else if (char === '}') braceDepth--;
                else if (char === '(') parenDepth++;
                else if (char === ')') parenDepth--;
            }
            
            // Split on semicolon only at top level
            if (char === ';' && !inString && braceDepth === 0 && parenDepth === 0) {
                if (current.trim()) {
                    statements.push(current.trim());
                }
                current = '';
            } else {
                current += char;
            }
        }
        
        // Don't forget the last statement
        if (current.trim()) {
            statements.push(current.trim());
        }
        
        return statements;
    },

    /**
     * Parses p-scope expression to identify output variables (those being assigned)
     */
    getPScopeOutputVars(pScopeExpr) {
        const outputVars = [];
        const statements = this.splitPScopeStatements(pScopeExpr);
        
        statements.forEach(stmt => {
            // Array destructuring: [a, b] = expr
            const arrayMatch = stmt.match(/^\[([^\]]+)\]\s*=\s*(.+)$/);
            if (arrayMatch) {
                const vars = arrayMatch[1].split(',').map(v => v.trim());
                outputVars.push(...vars);
                return;
            }
            
            // Object destructuring: {a, b} = expr
            const objMatch = stmt.match(/^\{([^}]+)\}\s*=\s*(.+)$/);
            if (objMatch) {
                const vars = objMatch[1].split(',').map(v => v.trim());
                outputVars.push(...vars);
                return;
            }
            
            // Regular variable assignment: var = expr
            const match = stmt.match(/^(\w+)\s*=\s*(.+)$/);
            if (match) {
                outputVars.push(match[1]);
            }
        });
        
        return outputVars;
    },

    /**
     * Creates a new scope for an element with p-scope during hydration
     */
    initScope(el, parentScope) {
        const dataId = el.getAttribute('p-id') || this.generateDomPathId(el);
        const localRawData = parentScope._p_target._p_children && parentScope._p_target._p_children[dataId];
        
        // Check for both p-scope and p-scope:sync on the same element
        const pScopeAttr = el.getAttribute('p-scope');
        const pScopeSyncAttr = Array.from(el.attributes).find(attr => attr.name.startsWith('p-scope') && attr.name.includes('sync'));
        
        // Collect all scope expressions and sync variables
        let allExprs = [];
        let syncVars = [];
        
        if (pScopeAttr) {
            // Regular p-scope (not synced)
            allExprs.push(pScopeAttr);
        }
        
        if (pScopeSyncAttr) {
            // p-scope:sync - these variables are synced
            const syncExpr = el.getAttribute(pScopeSyncAttr.name);
            allExprs.push(syncExpr);
            syncVars = this.getPScopeOutputVars(syncExpr);
        }
        
        // Combine all expressions
        const pScopeExpr = allExprs.join('; ');
        
        // Create new inherited Proxy with sync configuration
        const scope = this.observe(localRawData, parentScope, syncVars);
        
        // Store sync info on element for later use
        if (syncVars.length > 0) {
            el._p_syncVars = new Set(syncVars);
        }
        
        // Execute p-scope assignments directly on target to avoid triggering setter.
        // Pass the sync vars so executePScopeStatements can shadow getter/setter objects
        // assigned to non-sync props (preventing setter write-through to parent scope).
        this.executePScopeStatements(scope, pScopeExpr, new Set(syncVars));
        
        // Initialize parent snapshot so first refresh doesn't think everything changed
        const parentProto = Object.getPrototypeOf(scope._p_target);
        el._parentSnapshot = {};
        for (let key in parentProto) {
            if (!key.startsWith('_p_')) {
                el._parentSnapshot[key] = parentProto[key];
            }
        }
        
        // Initialize local snapshot for tracking local variable changes
        const target = scope._p_target;
        el._localSnapshot = {};
        for (let key of Object.keys(target)) {
            if (!key.startsWith('_p_')) {
                el._localSnapshot[key] = target[key];
            }
        }
        
        return scope;
    },

    /**
     * Gets the stored scope for an element during refresh, re-executing p-scope if parent changed
     */
    refreshScope(el, parentScope) {
        let scope = el._scope;
        
        if (!scope) {
            return parentScope;
        }
        
        // Check for both p-scope and p-scope:sync on the same element
        const pScopeAttr = el.getAttribute('p-scope');
        const pScopeSyncAttr = Array.from(el.attributes).find(attr => attr.name.startsWith('p-scope') && attr.name.includes('sync'));
        
        // Collect all scope expressions (just like in initScope)
        let allExprs = [];
        
        if (pScopeAttr) {
            allExprs.push(pScopeAttr);
        }
        
        if (pScopeSyncAttr) {
            allExprs.push(el.getAttribute(pScopeSyncAttr.name));
        }
        
        const pScopeExpr = allExprs.join('; ');
        
        // Check if parent values changed - if so, selectively re-execute p-scope
        if (pScopeExpr && scope._p_target) {
            this.updateScopeFromParent(el, scope, pScopeExpr);
        }
        
        return scope;
    },

    /**
     * Parses and executes p-scope statements sequentially, setting values directly on target.
     * Each statement sees the results of previous statements.
     *
     * syncVarsSet: Set of variable names that are sync props (p-scope:sync). Getter/setter
     * objects assigned to non-sync vars are wrapped in a shadow to prevent setter write-through.
     */
    executePScopeStatements(scope, pScopeExpr, syncVarsSet = new Set()) {
        const statements = this.splitPScopeStatements(pScopeExpr);
        const target = scope._p_target;
        
        // Helper: apply getter/setter shadow for non-sync object assignments.
        // ONLY shadow when the value is the exact same object reference already
        // present in the parent scope (i.e. it was passed down as a prop via
        // "content = content"). Objects *created* in this scope (like root-level
        // getter/setter proxy objects "content = { get name()... }") must NOT be
        // shadowed — their setters reference variables in THIS scope, so shadowing
        // would prevent reactive updates in the current scope.
        const parentTarget = Object.getPrototypeOf(target);
        const maybeShade = (varName, value) => {
            if (!syncVarsSet.has(varName) && value !== null && typeof value === 'object') {
                const comesFromParent = parentTarget && value === parentTarget[varName];
                if (comesFromParent) {
                    const shadow = this._createGetterSetterShadow(value);
                    if (shadow) return shadow;
                }
            }
            return value;
        };

        // Create a sequential scope that always reads from target first (for updated values)
        // then falls back to the parent scope for inherited values.
        //
        // The set trap is critical for getter/setter objects defined IN this scope
        // (e.g. `get_set_vars = {get name(){return name;}, set name(v){name=v;}}`).
        // Those setter functions close over this sequentialScope via `with`.  Without
        // a set trap, calling the setter later writes directly to `target` and bypasses
        // the reactive proxy, so pattr never learns that a value changed.  Routing
        // writes through `scope` (the reactive proxy) ensures walkDom is triggered.
        const sequentialScope = new Proxy(target, {
            get: (t, key) => {
                // First check if we have a local value (possibly updated by previous statement)
                if (Object.prototype.hasOwnProperty.call(t, key)) {
                    return t[key];
                }
                // Fall back to parent scope via prototype chain or original scope
                return scope[key];
            },
            set: (t, key, value) => {
                // Forward through the reactive proxy so setter functions defined inside
                // p-scope expressions trigger walkDom when invoked post-setup.
                scope[key] = value;
                return true;
            },
            has: (t, key) => {
                return Object.prototype.hasOwnProperty.call(t, key) || key in scope;
            }
        });
        
        for (const stmt of statements) {
            // Check for array destructuring: [a, b] = expr
            const arrayDestructMatch = stmt.match(/^\[([^\]]+)\]\s*=\s*(.+)$/);
            if (arrayDestructMatch) {
                const [, vars, expr] = arrayDestructMatch;
                const varNames = vars.split(',').map(v => v.trim());
                try {
                    const value = eval(`with (sequentialScope) { (${expr}) }`);
                    if (Array.isArray(value)) {
                        varNames.forEach((varName, i) => {
                            target[varName] = maybeShade(varName, value[i]);
                        });
                    }
                } catch (e) {
                    console.error(`Error executing p-scope array destructuring "${stmt}":`, e);
                }
                continue;
            }
            
            // Check for object destructuring: {a, b} = expr
            const objDestructMatch = stmt.match(/^\{([^}]+)\}\s*=\s*(.+)$/);
            if (objDestructMatch) {
                const [, vars, expr] = objDestructMatch;
                const varNames = vars.split(',').map(v => v.trim());
                try {
                    const value = eval(`with (sequentialScope) { (${expr}) }`);
                    if (typeof value === 'object' && value !== null) {
                        varNames.forEach(varName => {
                            target[varName] = maybeShade(varName, value[varName]);
                        });
                    }
                } catch (e) {
                    console.error(`Error executing p-scope object destructuring "${stmt}":`, e);
                }
                continue;
            }
            
            // Regular variable assignment
            const match = stmt.match(/^(\w+)\s*=\s*(.+)$/);
            if (match) {
                const [, varName, expr] = match;
                try {
                    const value = eval(`with (sequentialScope) { (${expr}) }`);
                    target[varName] = maybeShade(varName, value);
                } catch (e) {
                    console.error(`Error executing p-scope statement "${stmt}":`, e);
                }
            }
        }
    },

    /**
     * Re-executes p-scope statements that depend on changed variables (parent or local)
     * Statements are executed sequentially so each sees results of previous ones
     */
    updateScopeFromParent(el, scope, pScopeExpr) {
        const parentProto = Object.getPrototypeOf(scope._p_target);
        const target = scope._p_target;
        
        // Track which variables changed in PARENT vs LOCAL separately
        const changedParentVars = new Set();
        const changedLocalVars = new Set();
        
        // Check parent variable changes
        if (!el._parentSnapshot) {
            el._parentSnapshot = {};
        }
        for (let key in parentProto) {
            if (!key.startsWith('_p_')) {
                if (el._parentSnapshot[key] !== parentProto[key]) {
                    changedParentVars.add(key);
                }
                el._parentSnapshot[key] = parentProto[key];
            }
        }

        // For non-sync shadow objects: when a parent variable changes, clear any
        // matching explicit child overrides so the getter reads through to the
        // parent's new value.  This enforces one-way p-scope semantics:
        //   parent re-changes always propagate to the child, even if the child
        //   had previously set a local value for that same property.
        if (changedParentVars.size > 0) {
            const scopeSyncVarsSet = el._p_syncVars || new Set();
            for (const ownKey of Object.keys(target)) {
                if (ownKey.startsWith('_p_') || scopeSyncVarsSet.has(ownKey)) continue;
                const val = target[ownKey];
                if (val && typeof val === 'object' && val._p_isShadow) {
                    for (const changedVar of changedParentVars) {
                        if (val._p_explicitOverrideKeys.has(changedVar)) {
                            delete val._p_localOverrides[changedVar];
                            val._p_explicitOverrideKeys.delete(changedVar);
                        }
                    }
                }
            }
        }
        
        // Check local variable changes (variables that are OWN properties of target)
        if (!el._localSnapshot) {
            el._localSnapshot = {};
        }
        for (let key of Object.keys(target)) {
            if (!key.startsWith('_p_')) {
                if (el._localSnapshot[key] !== target[key]) {
                    changedLocalVars.add(key);
                }
                el._localSnapshot[key] = target[key];
            }
        }
        
        // Combine for checking which statements to execute
        const allChangedVars = new Set([...changedParentVars, ...changedLocalVars]);
        
        if (allChangedVars.size === 0) return;
        
        try {
            const statements = this.splitPScopeStatements(pScopeExpr);
            
            // Identify OUTPUT variables (variables SET by p-scope statements)
            // Local changes to these should NOT trigger re-execution
            const outputVars = new Set();
            statements.forEach(stmt => {
                // Handle array destructuring: [a, b] = expr
                const arrayMatch = stmt.match(/^\[([^\]]+)\]\s*=\s*(.+)$/);
                if (arrayMatch) {
                    const vars = arrayMatch[1].split(',').map(v => v.trim());
                    vars.forEach(v => outputVars.add(v));
                    return;
                }
                
                // Handle object destructuring: {a, b} = expr
                const objMatch = stmt.match(/^\{([^}]+)\}\s*=\s*(.+)$/);
                if (objMatch) {
                    const vars = objMatch[1].split(',').map(v => v.trim());
                    vars.forEach(v => outputVars.add(v));
                    return;
                }
                
                // Regular variable
                const match = stmt.match(/^(\w+)\s*=\s*(.+)$/);
                if (match) {
                    outputVars.add(match[1]);
                }
            });
            
            // Track variables set during THIS re-execution pass
            const setInThisPass = new Set();
            
            // Create a sequential scope that:
            // - Uses values from this pass if already computed
            // - For PARENT-changed vars: reads from parent (new value)
            // - For LOCAL-changed vars: reads from local (new value)
            // - Otherwise: reads from local if exists, else parent
            const scopeSyncVars = el._p_syncVars || new Set();
            const sequentialScope = new Proxy(target, {
                get: (t, key) => {
                    if (key === '_p_target' || key === '_p_children' || key === '_p_scope') {
                        return t[key];
                    }
                    // If this variable was set by a previous statement in this pass, use that
                    if (setInThisPass.has(key)) {
                        return t[key];
                    }
                    // If this variable changed in PARENT, read from parent (new value)
                    if (changedParentVars.has(key)) {
                        return parentProto[key];
                    }
                    // If this variable changed LOCALLY, read from local (new value)
                    if (changedLocalVars.has(key)) {
                        return t[key];
                    }
                    // Otherwise, read current value (local if exists, else parent)
                    if (Object.prototype.hasOwnProperty.call(t, key)) {
                        return t[key];
                    }
                    return parentProto[key];
                },
                set: (t, key, value) => {
                    // Apply the same getter/setter shadow logic as executePScopeStatements
                    // (comesFromParent guard + preserve ONLY explicitly child-written overrides
                    // so that parent mutations to arrays propagate correctly on re-shadow).
                    if (!scopeSyncVars.has(key) && value !== null && typeof value === 'object') {
                        const parentProto2 = Object.getPrototypeOf(t);
                        const comesFromParent = parentProto2 && value === parentProto2[key];
                        if (comesFromParent) {
                            const existing = Object.prototype.hasOwnProperty.call(t, key) ? t[key] : null;
                            const existingOverrides = existing && existing._p_isShadow ? existing._p_localOverrides : null;
                            const existingExplicitKeys = existing && existing._p_isShadow ? existing._p_explicitOverrideKeys : null;
                            const shadow = this._createGetterSetterShadow(value, existingOverrides, existingExplicitKeys);
                            if (shadow) value = shadow;
                        }
                    }
                    t[key] = value;
                    return true;
                },
                has: (t, key) => {
                    return setInThisPass.has(key) || Object.prototype.hasOwnProperty.call(t, key) || key in parentProto;
                }
            });
            
            statements.forEach(stmt => {
                let shouldExecute = false;
                const parts = stmt.split('=');
                if (parts.length <= 1) return;
                const rhs = parts.slice(1).join('=');
                
                // Execute if RHS contains a PARENT-changed variable
                changedParentVars.forEach(varName => {
                    if (rhs.includes(varName)) {
                        shouldExecute = true;
                    }
                });
                
                // Execute if RHS contains a LOCAL-changed variable that is NOT an output var
                // (Local changes to output vars are user modifications that should stick)
                changedLocalVars.forEach(varName => {
                    if (!outputVars.has(varName) && rhs.includes(varName)) {
                        shouldExecute = true;
                    }
                });
                
                // Also execute if the statement depends on a variable set earlier in this pass
                if (!shouldExecute) {
                    setInThisPass.forEach(varName => {
                        if (rhs.includes(varName)) {
                            shouldExecute = true;
                        }
                    });
                }
                
                if (shouldExecute) {
                    // Handle array destructuring: [a, b] = expr
                    const arrayMatch = stmt.match(/^\[([^\]]+)\]\s*=\s*(.+)$/);
                    if (arrayMatch) {
                        const [, vars, expr] = arrayMatch;
                        const varNames = vars.split(',').map(v => v.trim());
                        const value = eval(`with (sequentialScope) { (${expr}) }`);
                        if (Array.isArray(value)) {
                            varNames.forEach((varName, i) => {
                                target[varName] = value[i];
                                setInThisPass.add(varName);
                            });
                        }
                        return;
                    }
                    
                    // Handle object destructuring: {a, b} = expr
                    const objMatch = stmt.match(/^\{([^}]+)\}\s*=\s*(.+)$/);
                    if (objMatch) {
                        const [, vars, expr] = objMatch;
                        const varNames = vars.split(',').map(v => v.trim());
                        const value = eval(`with (sequentialScope) { (${expr}) }`);
                        if (typeof value === 'object' && value !== null) {
                            varNames.forEach(varName => {
                                target[varName] = value[varName];
                                setInThisPass.add(varName);
                            });
                        }
                        return;
                    }
                    
                    // Regular variable assignment
                    const match = stmt.match(/^(\w+)\s*=\s*(.+)$/);
                    if (match) {
                        const [, varName, expr] = match;
                        let value = eval(`with (sequentialScope) { (${expr}) }`);
                        // Mirror the comesFromParent shadow logic from executePScopeStatements.
                        // Preserve ONLY explicitly child-written overrides so that parent
                        // mutations to arrays propagate correctly on re-shadow.
                        if (!scopeSyncVars.has(varName) && value !== null && typeof value === 'object') {
                            const comesFromParent = parentProto && value === parentProto[varName];
                            if (comesFromParent) {
                                const existing = Object.prototype.hasOwnProperty.call(target, varName) ? target[varName] : null;
                                const existingOverrides = existing && existing._p_isShadow ? existing._p_localOverrides : null;
                                const existingExplicitKeys = existing && existing._p_isShadow ? existing._p_explicitOverrideKeys : null;
                                const shadow = this._createGetterSetterShadow(value, existingOverrides, existingExplicitKeys);
                                if (shadow) value = shadow;
                            }
                        }
                        target[varName] = value;
                        setInThisPass.add(varName);
                    }
                }
            });
            
            // Update local snapshot with any newly computed values
            for (let key of Object.keys(target)) {
                if (!key.startsWith('_p_')) {
                    el._localSnapshot[key] = target[key];
                }
            }
        } catch (e) {
            console.error(`Error re-executing p-scope expression:`, e);
        }
    },

    // ==================== EVENT HANDLING ====================

    /**
     * Registers p-on:* event listeners on an element
     */
    registerEventListeners(el) {
        Array.from(el.attributes).forEach(attr => {
            if (attr.name.startsWith('p-on:')) {
                const event = attr.name.replace('p-on:', '');
                el.addEventListener(event, () => {
                    eval(`with (el._scope) { ${attr.value} }`);
                    this.walkDom(this.root, this.data, false);
                    this.refreshAllLoops();
                });
            }
        });
    },

    /**
     * Sets up p-model two-way binding on an element
     */
    registerModelBinding(el) {
        const modelAttr = el.getAttribute('p-model');
        if (!modelAttr) return;
        
        el.addEventListener('input', (e) => {
            let value;
            const type = e.target.type;
            
            if (type === 'number' || type === 'range') {
                value = e.target.value === '' ? null : Number(e.target.value);
            } else if (type === 'checkbox') {
                value = e.target.checked;
            } else if (type === 'radio') {
                value = e.target.checked ? e.target.value : undefined;
            } else {
                // Check if current value is an array
                const currentValue = eval(`with (el._scope) { ${modelAttr} }`);
                if (Array.isArray(currentValue)) {
                    // Split by comma and trim, filter out empty strings
                    value = e.target.value.split(',').map(s => s.trim()).filter(s => s);
                } else {
                    value = e.target.value;
                }
            }
            
            if (value !== undefined) {
                // Set on the scope (Proxy) to trigger reactive updates
                // This handles cases like content[key] where content is inherited
                const scope = el._scope;
                
                // For bracket notation like content[key] or nested content[key][i],
                // split at the LAST bracket pair: evaluate the container object and the key
                // separately through the scope chain, then set directly on the container.
                // This handles arbitrary nesting depth AND avoids a subtle `with`-statement
                // variable-shadowing bug: if the loop uses `for [key, value] of ...` then
                // `value` is a scope property, so `eval('with (scope) { expr = value }')` would
                // assign the loop variable back to itself instead of the new input value.
                if (modelAttr.includes('[')) {
                    const lastBracket = modelAttr.lastIndexOf('[');
                    const containerExpr = modelAttr.substring(0, lastBracket);
                    const keyExpr = modelAttr.substring(lastBracket + 1, modelAttr.length - 1);
                    const container = eval(`with (scope) { (${containerExpr}) }`);
                    const key = eval(`with (scope) { (${keyExpr}) }`);
                    container[key] = value;
                } else if (modelAttr.includes('.')) {
                    // Handle dot notation like content.path
                    const parts = modelAttr.split('.');
                    let obj = scope;
                    for (let i = 0; i < parts.length - 1; i++) {
                        obj = obj[parts[i]];
                    }
                    obj[parts[parts.length - 1]] = value;
                } else {
                    scope[modelAttr] = value;  // Set on Proxy to trigger reactivity
                }
                
                // Track the input element to restore focus after re-render
                // Get p-for-key from element or its parent
                let pForKey = el.getAttribute('p-for-key');
                let container = el;
                if (!pForKey) {
                    container = el.parentElement;
                    while (container && !pForKey) {
                        pForKey = container.getAttribute('p-for-key');
                        if (!pForKey) container = container.parentElement;
                    }
                }
                
                // Store the p-model value to identify this specific input
                const modelAttrValue = el.getAttribute('p-model');
                const inputTagName = el.tagName;

                // Save cursor position before re-render so we can restore it after
                // the loop re-creates the input element.
                const selStart = el.selectionStart;
                const selEnd   = el.selectionEnd;
                
                // Re-render DOM to reflect changes
                this.walkDom(this.root, this.data, false);
                
                // Restore focus AND cursor position after DOM update.
                // Only needed for inputs inside p-for loops — those elements are
                // removed and re-created by refreshLoop, which drops focus and resets
                // the selection. Non-loop inputs are updated in-place, so they keep focus.
                if (pForKey && modelAttrValue) {
                    requestAnimationFrame(() => {
                        // First try: find the input directly by p-for-key + p-model.
                        // This handles "flat" loop templates where the input itself carries
                        // p-for-key (e.g. <div p-for-key> and <input p-for-key> are siblings).
                        // Note: modelAttrValue may contain [ ] characters (e.g. "obj[key]");
                        // inside a quoted CSS attribute value these are treated as literals.
                        let elementToFocus = document.querySelector(
                            `${inputTagName.toLowerCase()}[p-for-key="${pForKey}"][p-model="${modelAttrValue}"]`
                        );
                        // Fallback: find the container with p-for-key, then find input with matching p-model inside it.
                        // This handles the more common case where a wrapper div carries p-for-key.
                        if (!elementToFocus) {
                            const containerEl = document.querySelector(`[p-for-key="${pForKey}"]`);
                            if (containerEl) {
                                elementToFocus = containerEl.querySelector(`${inputTagName.toLowerCase()}[p-model="${modelAttrValue}"]`);
                            }
                        }
                        if (elementToFocus) {
                            elementToFocus.focus();
                            // Restore cursor/selection position.
                            // setSelectionRange is only valid on text-like inputs (not number, date, etc.)
                            if (selStart !== null && selEnd !== null) {
                                try {
                                    elementToFocus.setSelectionRange(selStart, selEnd);
                                } catch (_) { /* ignore for input types that don't support selection */ }
                            }
                        }
                    });
                }
            }
        });
        
        // For checkbox and radio, also listen to 'change' event
        if (el.type === 'checkbox' || el.type === 'radio') {
            el.addEventListener('change', (e) => {
                let value;
                if (el.type === 'checkbox') {
                    value = e.target.checked;
                } else {
                    value = e.target.checked ? e.target.value : undefined;
                }
            if (value !== undefined) {
                const scope = el._scope;
                if (modelAttr.includes('[')) {
                    const lastBracket = modelAttr.lastIndexOf('[');
                    const containerExpr = modelAttr.substring(0, lastBracket);
                    const keyExpr = modelAttr.substring(lastBracket + 1, modelAttr.length - 1);
                    const container = eval(`with (scope) { (${containerExpr}) }`);
                    const key = eval(`with (scope) { (${keyExpr}) }`);
                    container[key] = value;
                } else {
                    scope[modelAttr] = value;
                }
            }
            });
        }
    },

    // ==================== DIRECTIVE EVALUATION ====================

    /**
     * Evaluates all directives on an element
     */
    evaluateDirectives(el, scope) {
        Array.from(el.attributes).forEach(attr => {
            const parsed = this.parseDirectiveModifiers(attr.name);
            if (Object.keys(this.directives).includes(parsed.directive)) {
                const evalScope = el._scope || scope;
                const value = eval(`with (evalScope) { (${attr.value}) }`);
                this.directives[parsed.directive](el, value, parsed.modifiers);
            }
        });
    },

    // ==================== LOOP HANDLING ====================

    setForTemplateRecursive(element, template) {
        element._forTemplate = template;
        Array.from(element.children).forEach(child => {
            this.setForTemplateRecursive(child, template);
        });
    },

    refreshAllLoops(el = this.root, processedTemplates = new Set()) {
        if (el.tagName === 'TEMPLATE' && el.hasAttribute('p-for')) {
            // Skip if this template was already processed during this refresh cycle
            // (e.g., as a nested template during an outer loop's refresh)
            if (processedTemplates.has(el)) {
                return;
            }
            processedTemplates.add(el);
            this.handleFor(el, el._scope || this.data, false);
            return;
        }
        
        let child = el.firstElementChild;
        while (child) {
            const nextChild = child.nextElementSibling; // Store before processing in case DOM changes
            this.refreshAllLoops(child, processedTemplates);
            child = nextChild;
        }
    },

    handleFor(template, parentScope, isHydrating) {
        const forExpr = template.getAttribute('p-for');
        const match = forExpr.match(/^(?:const|let)?\s*(.+?)\s+(of|in)\s+(.+)$/);
        
        if (!match) {
            console.error(`Invalid p-for expression: ${forExpr}`);
            return;
        }
        
        const [, varPattern, operator, iterableExpr] = match;
        
        if (isHydrating) {
            this.hydrateLoop(template, parentScope, varPattern, iterableExpr);
        } else {
            this.refreshLoop(template, parentScope, varPattern, iterableExpr);
        }
    },

    // ==================== DYNAMIC COMPONENTS ====================

    /**
     * Looks up a component in the registry by name.
     * Registry entries are inert templates emitted by the backend (e.g. Pico):
     * <template p-comp-src="hero"> ...component markup with p-* bindings... </template>
     */
    findCompTemplate(name) {
        const registry = document.querySelectorAll('template[p-comp-src]');
        for (const tpl of registry) {
            if (tpl.getAttribute('p-comp-src') === name) return tpl;
        }
        return null;
    },

    /**
     * Handles <template p-comp="expr"> anchors (dynamic components).
     *
     * The template stays in the DOM as an inert anchor (same pattern as
     * <template p-for>); the resolved component's markup is cloned from the
     * registry and inserted as siblings right after the anchor, so no wrapper
     * element is introduced and direct-child CSS selectors keep working.
     *
     * Rendered nodes are tracked in template._compData.renderedElements so
     * removeLoopElements/findLastRenderedSibling can account for them.
     *
     * SSR contract: backend-rendered component markup carries a
     * p-comp-node="name" attribute on each top-level node. During hydration the
     * anchor adopts those siblings instead of re-inflating from the registry.
     * (Inside p-for loops the anchor itself must carry the item's p-for-key so
     * hydrateLoop's sibling grouping isn't interrupted.)
     */
    handleComp(template, parentScope, isHydrating) {
        const evalScope = template._scope || parentScope;
        template._scope = evalScope;

        let name;
        try {
            name = eval(`with (evalScope) { (${template.getAttribute('p-comp')}) }`);
        } catch (e) {
            console.error(`Error evaluating p-comp expression: ${template.getAttribute('p-comp')}`, e);
            return;
        }
        if (name !== null && name !== undefined) name = String(name);

        const compData = template._compData || (template._compData = { name: null, renderedElements: [] });

        // Already rendering this component and it's still in the DOM — nothing to do.
        // (Covers repeat refresh walks as well as the empty-name steady state.)
        if (compData.name === name && compData.renderedElements.every(el => el.isConnected)) {
            return;
        }

        // Tear down previously rendered content (component removed or swapped)
        if (compData.renderedElements.length > 0) {
            this.removeLoopElements(compData.renderedElements);
            compData.renderedElements = [];
        }
        compData.name = name;

        if (name === null || name === undefined || name === '') {
            return;
        }

        // SSR adoption: adopt following siblings marked p-comp-node="<name>".
        // They are hydrated by the normal walkDom pass, so don't walk them here.
        if (isHydrating) {
            let sibling = template.nextElementSibling;
            while (sibling && sibling.getAttribute('p-comp-node') === name) {
                compData.renderedElements.push(sibling);
                sibling = sibling.nextElementSibling;
            }
            if (compData.renderedElements.length > 0) {
                return;
            }
        }

        const src = this.findCompTemplate(name);
        if (!src) {
            console.error(`p-comp: no component registered under name "${name}" (expected a <template p-comp-src="${name}">)`);
            return;
        }

        const clone = src.content.cloneNode(true);
        const elements = Array.from(clone.children);

        // If this anchor belongs to a p-for loop, mark the inflated content with the
        // same ownership so refresh walks treat nested templates like loop clones
        // (the parent loop rebuilds them).
        if (template._forTemplate) {
            elements.forEach(el => this.setForTemplateRecursive(el, template._forTemplate));
        }

        // Stamp the owning loop item's p-for-key on the inflated content so
        // client-rendered output matches SSR markup and the p-model
        // focus-restoration lookup can find re-rendered inputs inside components.
        // Nested p-comp anchors inherit the key via _p_forKey (set before the walk
        // below inflates them); templates keep it as a JS property only, matching
        // the loop render branches which never stamp key attributes on templates.
        const forKey = template._p_forKey || template.getAttribute('p-for-key');
        if (forKey) {
            elements.forEach(el => {
                el._p_forKey = forKey;
                if (el.tagName !== 'TEMPLATE') {
                    el.setAttribute('p-for-key', forKey);
                }
                el.querySelectorAll('template[p-comp]').forEach(tpl => { tpl._p_forKey = forKey; });
            });
        }

        const fragment = document.createDocumentFragment();
        elements.forEach(el => fragment.appendChild(el));
        template.parentNode.insertBefore(fragment, template.nextSibling);

        compData.renderedElements = elements;

        // Hydrate the new nodes (scopes, directives, event listeners, model
        // bindings, nested p-for/p-comp templates).
        elements.forEach(el => this.walkDom(el, evalScope, true));
    },

    /**
     * Gets the scope prefix for a template based on its nesting level
     * Checks ancestors and siblings for parent p-for templates to build hierarchical key
     */
    getTemplateScopePrefix(template) {
        let ancestorKey = '';
        
        // First, check if SSR elements exist after this template
        // If so, extract the scope ID from them to maintain continuity
        let nextSibling = template.nextElementSibling;
        if (nextSibling && nextSibling.hasAttribute && nextSibling.hasAttribute('p-for-key')) {
            const ssrKey = nextSibling.getAttribute('p-for-key');
            // Extract scope prefix from SSR element (e.g., "s4:0" -> "s4:")
            const colonIndex = ssrKey.lastIndexOf(':');
            if (colonIndex > 0) {
                const ssrScopePrefix = ssrKey.substring(0, colonIndex + 1);
                // Extract just the scope ID part (e.g., "s4:" or "s3:0-s4:")
                const lastDashIndex = ssrScopePrefix.lastIndexOf('-');
                if (lastDashIndex >= 0) {
                    // Has ancestor prefix, extract both parts
                    ancestorKey = ssrScopePrefix.substring(0, lastDashIndex + 1);
                    const scopeId = ssrScopePrefix.substring(lastDashIndex + 1, ssrScopePrefix.length - 1);
                    template._forScopeId = scopeId;
                } else {
                    // No ancestor prefix, just scope ID
                    template._forScopeId = ssrScopePrefix.substring(0, ssrScopePrefix.length - 1);
                }
                // Persist the scope ID as an attribute so it survives cloneNode.
                // When refreshLoop clones the outer template's content, the inner template
                // clone carries this attribute and reuses the same ID on the next render,
                // keeping p-for-key values stable across re-renders.
                template.setAttribute('p-for-scope-id', template._forScopeId);
                return ssrScopePrefix;
            }
        }
        
        // Check if this template has _forTemplate set (meaning it was rendered by an outer loop)
        if (template._forTemplate) {
            const parentForData = template._forTemplate._forData;
            if (parentForData && parentForData.scopePrefix) {
                // Find which iteration we're in by checking sibling elements
                let sibling = template.previousElementSibling;
                while (sibling) {
                    if (sibling.hasAttribute && sibling.hasAttribute('p-for-key')) {
                        const siblingKey = sibling.getAttribute('p-for-key');
                        ancestorKey = siblingKey + '-';
                        break;
                    }
                    sibling = sibling.previousElementSibling;
                }
            }
        }
        
        // If no _forTemplate, check previous siblings for p-for-key
        // This handles the case where template is adjacent to loop-rendered elements
        if (!ancestorKey) {
            let sibling = template.previousElementSibling;
            while (sibling) {
                if (sibling.hasAttribute && sibling.hasAttribute('p-for-key')) {
                    // Check if this sibling shares the same parent template
                    if (sibling._forTemplate === template._forTemplate) {
                        const siblingKey = sibling.getAttribute('p-for-key');
                        ancestorKey = siblingKey + '-';
                        break;
                    }
                }
                sibling = sibling.previousElementSibling;
            }
        }
        
        // Check parent elements for p-for-key
        if (!ancestorKey) {
            let parent = template.parentElement;
            while (parent) {
                if (parent.hasAttribute && parent.hasAttribute('p-for-key')) {
                    const parentKey = parent.getAttribute('p-for-key');
                    ancestorKey = parentKey + '-';
                    break;
                }
                if (parent._forTemplate) {
                    const parentForData = parent._forTemplate._forData;
                    if (parentForData) {
                        const parentIndex = parentForData.renderedElements.indexOf(parent);
                        if (parentIndex >= 0) {
                            ancestorKey = (parentForData.scopePrefix || '') + parentIndex + '-';
                            break;
                        }
                    }
                }
                parent = parent.parentElement;
            }
        }
        
        // Generate a unique scope ID for this template if not already set from SSR.
        // Also check p-for-scope-id: cloneNode copies HTML attributes but not JS
        // properties, so a cloned inner template will have the attribute but not _forScopeId.
        // Reading it back here keeps the same ID across refreshLoop re-renders.
        if (!template._forScopeId) {
            if (template.hasAttribute('p-for-scope-id')) {
                // Restore the ID that was persisted by a previous call on the original template
                template._forScopeId = template.getAttribute('p-for-scope-id');
            } else {
                template._forScopeId = 's' + (this._templateScopeCounter++);
                // Persist so future clones of this template also reuse the same ID
                template.setAttribute('p-for-scope-id', template._forScopeId);
            }
        }
        
        return ancestorKey + template._forScopeId + ':';
    },

    /**
     * Recursively enriches a template content element with nested p-for templates
     * found in SSR-rendered elements but missing from the outer template's static content.
     *
     * When a template generator (e.g. Pico) emits a nested {for} loop only inside the
     * SSR-rendered siblings—but NOT inside the outer <template p-for> content—refreshLoop
     * will re-render from template.content and lose the inner template entirely.
     * Calling this once per hydrateLoop pass "teaches" the outer template about the inner
     * template so all future refreshLoop calls reproduce it correctly.
     */
    enrichTemplateContent(templateEl, ssrEl) {
        // Filter out SSR-rendered loop elements (those with p-for-key)
        const ssrChildren = Array.from(ssrEl.children).filter(
            c => !c.hasAttribute('p-for-key')
        );
        const tplChildren = Array.from(templateEl.children);
        let ti = 0;

        for (const ssrChild of ssrChildren) {
            if (ssrChild.tagName === 'TEMPLATE' && ssrChild.hasAttribute('p-for')) {
                // Check whether the template content already has a matching inner template here
                const tplChild = tplChildren[ti];
                const alreadyPresent =
                    tplChild &&
                    tplChild.tagName === 'TEMPLATE' &&
                    tplChild.getAttribute('p-for') === ssrChild.getAttribute('p-for');

                let targetTpl; // the template element inside outerTemplate.content to annotate
                if (!alreadyPresent) {
                    // Insert a clone of the SSR inner template into the outer template content.
                    // cloneNode(true) copies the <template> element and its .content fragment
                    // but does NOT copy the SSR-rendered siblings (p-for-key elements).
                    const clonedInnerTemplate = ssrChild.cloneNode(true);
                    templateEl.insertBefore(clonedInnerTemplate, tplChild || null);
                    tplChildren.splice(ti, 0, clonedInnerTemplate);
                    targetTpl = clonedInnerTemplate;
                } else {
                    targetTpl = tplChild;
                }

                // Stamp a stable p-for-scope-id on the template inside outerTemplate.content
                // so that every subsequent refreshLoop clone of that content inherits the same
                // scope ID.  Without this, each clone starts without _forScopeId (cloneNode
                // doesn't copy JS properties) and getTemplateScopePrefix mints a fresh counter
                // ID, making p-for-key values change on every re-render and breaking the
                // requestAnimationFrame focus-restoration querySelector.
                //
                // enrichTemplateContent runs *before* hydrateLoop processes SSR elements, but
                // the SSR-rendered loop items (div.array-item[p-for-key="s0:2-s4:0"] etc.) are
                // already in the DOM as siblings of ssrChild, so we can read the scope ID from
                // them right here.
                if (targetTpl && !targetTpl.hasAttribute('p-for-scope-id')) {
                    const firstLoopItem = ssrChild.nextElementSibling;
                    if (firstLoopItem && firstLoopItem.hasAttribute('p-for-key')) {
                        // Extract the inner scope ID segment from the SSR key
                        // e.g. "s0:2-s4:0" → scopePart "s0:2-s4" → scopeId "s4"
                        const ssrKey = firstLoopItem.getAttribute('p-for-key');
                        const col = ssrKey.lastIndexOf(':');
                        if (col > 0) {
                            const scopePart = ssrKey.substring(0, col);
                            const dash = scopePart.lastIndexOf('-');
                            const scopeId = dash >= 0 ? scopePart.substring(dash + 1) : scopePart;
                            targetTpl.setAttribute('p-for-scope-id', scopeId);
                        }
                    } else {
                        // No SSR items (e.g. initially-empty array) — generate a stable ID now
                        // so that all future refreshLoop clones also use the same ID.
                        targetTpl.setAttribute('p-for-scope-id', 's' + (this._templateScopeCounter++));
                    }
                }

                ti++;
            } else if (ti < tplChildren.length && tplChildren[ti].tagName === ssrChild.tagName) {
                // Structurally matching element — recurse into children
                this.enrichTemplateContent(tplChildren[ti], ssrChild);
                ti++;
            }
        }
    },

    hydrateLoop(template, parentScope, varPattern, iterableExpr) {
        template._scope = parentScope;
        
        try {
            const iterable = eval(`with (parentScope) { (${iterableExpr}) }`);
            
            // Get the scope prefix for this template's keys
            const scopePrefix = this.getTemplateScopePrefix(template);
            
            template._forData = {
                varPattern,
                iterableExpr,
                renderedElements: [],
                scopePrefix
            };
            
            // Check for SSR-rendered elements - only match those with our scope prefix
            const existingElementsByKey = {};
            let sibling = template.nextElementSibling;
            while (sibling && sibling.hasAttribute('p-for-key')) {
                const fullKey = sibling.getAttribute('p-for-key');
                
                // Check if this element belongs to this template (has our scope prefix)
                // or is an old-style unscoped key (for backwards compatibility)
                const isOurElement = fullKey.startsWith(scopePrefix) || 
                    (!fullKey.includes(':') && !fullKey.includes('-')); // unscoped legacy key
                
                if (isOurElement) {
                    // Extract the index part from the key
                    let indexKey;
                    if (fullKey.startsWith(scopePrefix)) {
                        indexKey = fullKey.substring(scopePrefix.length);
                    } else {
                        indexKey = fullKey; // legacy unscoped key
                    }
                    
                    if (!existingElementsByKey[indexKey]) {
                        existingElementsByKey[indexKey] = [];
                    }
                    existingElementsByKey[indexKey].push(sibling);
                }
                sibling = sibling.nextElementSibling;
            }

            // Enrich the outer template's content with any inner p-for templates found in
            // SSR-rendered siblings but absent from template.content.  Without this,
            // refreshLoop re-renders from template.content and loses the inner template.
            const templateRoot = template.content.firstElementChild;
            if (templateRoot) {
                for (const elements of Object.values(existingElementsByKey)) {
                    for (const el of elements) {
                        if (el.querySelector && el.querySelector('template[p-for]')) {
                            this.enrichTemplateContent(templateRoot, el);
                        }
                    }
                }
            }
            
            let index = 0;
            let lastInserted = template;
            
            for (const item of iterable) {
                const loopScope = this.createLoopScope(parentScope, varPattern, item);
                
                if (existingElementsByKey[String(index)]) {
                    // Hydrate existing SSR elements
                    existingElementsByKey[String(index)].forEach(el => {
                        el._scope = loopScope;
                        this.setForTemplateRecursive(el, template);
                        el._p_forKey = scopePrefix + index;
                        // Update key to use scoped format
                        if (el.tagName !== 'TEMPLATE') {
                            el.setAttribute('p-for-key', scopePrefix + index);
                        }
                        this.walkDom(el, loopScope, true);
                        // Mark AFTER walkDom so that the parent walkDom (which may still be
                        // iterating over children that include this element) knows to skip it
                        // and avoid re-evaluating its directives with the wrong outer scope.
                        el._p_loopHydrated = true;
                        template._forData.renderedElements.push(el);
                        lastInserted = el;
                    });
                } else {
                    // Render from template
                    const clone = template.content.cloneNode(true);
                    const elements = Array.from(clone.children);

                    elements.forEach(el => {
                        el._scope = loopScope;
                        this.setForTemplateRecursive(el, template);
                        el._p_forKey = scopePrefix + index;
                        // Only set p-for-key on non-template elements at this level
                        if (el.tagName !== 'TEMPLATE') {
                            el.setAttribute('p-for-key', scopePrefix + index);
                        }
                    });

                    // Insert elements after the last inserted element
                    const fragment = document.createDocumentFragment();
                    elements.forEach(el => fragment.appendChild(el));
                    lastInserted.parentNode.insertBefore(fragment, lastInserted.nextSibling);

                    template._forData.renderedElements.push(...elements);

                    // Walk AFTER insertion so <template p-comp> anchors are connected
                    // and can inflate their component content as siblings.
                    elements.forEach(el => {
                        this.walkDom(el, loopScope, true);
                    });

                    if (elements.length > 0) {
                        lastInserted = this.findLastRenderedSibling(elements[elements.length - 1]);
                    }
                }
                
                index++;
            }
            
            // Remove extra SSR elements
            Object.keys(existingElementsByKey).forEach(key => {
                if (parseInt(key) >= index) {
                    existingElementsByKey[key].forEach(el => el.remove());
                }
            });
        } catch (e) {
            console.error(`Error in p-for hydration: ${iterableExpr}`, e);
        }
    },

    /**
     * Recursively removes elements and their nested loop contents
     */
    removeLoopElements(elements) {
        elements.forEach(el => {
            // Remove any component content this p-comp anchor rendered
            if (el._compData && el._compData.renderedElements) {
                this.removeLoopElements(el._compData.renderedElements);
                el._compData.renderedElements = [];
                el._compData.name = null;
            }
            // If this element is a template with its own rendered elements, remove those first
            if (el._forData && el._forData.renderedElements) {
                this.removeLoopElements(el._forData.renderedElements);
                el._forData.renderedElements = [];
            }
            // Also check children for templates with rendered elements
            if (el.querySelectorAll) {
                const nestedTemplates = el.querySelectorAll('template[p-for]');
                nestedTemplates.forEach(tpl => {
                    if (tpl._forData && tpl._forData.renderedElements) {
                        this.removeLoopElements(tpl._forData.renderedElements);
                        tpl._forData.renderedElements = [];
                    }
                });
                const nestedCompTemplates = el.querySelectorAll('template[p-comp]');
                nestedCompTemplates.forEach(tpl => {
                    if (tpl._compData && tpl._compData.renderedElements) {
                        this.removeLoopElements(tpl._compData.renderedElements);
                        tpl._compData.renderedElements = [];
                        tpl._compData.name = null;
                    }
                });
            }
            el.remove();
        });
    },

    refreshLoop(template, parentScope, varPattern, iterableExpr) {
        const forData = template._forData;
        if (!forData) return;
        
        try {
            const iterable = eval(`with (parentScope._p_target || parentScope) { (${iterableExpr}) }`);

            // Materialise the iterable so we can (a) compare with the previous render
            // and (b) iterate it without risk of exhausting a one-pass generator.
            const items = iterable == null ? [] : Array.from(iterable);

            // Skip re-rendering if the loop data hasn't actually changed.
            // This prevents unnecessary DOM thrashing when a parent event handler
            // (e.g. p-on:blur/focus) triggers a walkDom whose visible loop data is
            // unchanged.  Array items (e.g. Object.entries pairs) are compared
            // element-by-element; primitives and object references use strict equality.
            const prevItems = forData._prevItems;
            if (prevItems && prevItems.length === items.length) {
                let changed = false;
                for (let i = 0; i < items.length && !changed; i++) {
                    const prev = prevItems[i];
                    const curr = items[i];
                    if (Array.isArray(prev) && Array.isArray(curr)) {
                        if (prev.length !== curr.length) {
                            changed = true;
                        } else {
                            for (let j = 0; j < prev.length && !changed; j++) {
                                if (prev[j] !== curr[j]) changed = true;
                            }
                        }
                    } else if (prev !== curr) {
                        changed = true;
                    }
                }
                if (!changed) return;
            }
            forData._prevItems = items;

            // Use stored scope prefix or regenerate if needed
            const scopePrefix = forData.scopePrefix || this.getTemplateScopePrefix(template);
            
            // Remove all elements (including nested loop elements) and re-render
            this.removeLoopElements(forData.renderedElements);
            forData.renderedElements = [];
            
            // Track the actual last inserted element (accounts for nested template output)
            let lastInsertedElement = template;
            
            let index = 0;
            for (const item of items) {
                const clone = template.content.cloneNode(true);
                const loopScope = this.createLoopScope(parentScope, forData.varPattern, item);
                
                const elements = Array.from(clone.children);
                
                // First pass: set up scope and basic attributes (but don't process nested templates yet)
                elements.forEach(el => {
                    el._scope = loopScope;
                    this.setForTemplateRecursive(el, template);
                    el._p_forKey = scopePrefix + index;
                    // Only set p-for-key on non-template elements
                    if (el.tagName !== 'TEMPLATE') {
                        el.setAttribute('p-for-key', scopePrefix + index);
                    }
                });
                
                // Insert elements into DOM after the last inserted element
                const fragment = document.createDocumentFragment();
                elements.forEach(el => fragment.appendChild(el));
                lastInsertedElement.parentNode.insertBefore(fragment, lastInsertedElement.nextSibling);
                forData.renderedElements.push(...elements);
                
                // Second pass: now that elements are in DOM, process them (including nested templates)
                elements.forEach(el => {
                    this.walkDom(el, loopScope, true);
                });
                
                // Update lastInsertedElement to the actual last element in the DOM from this iteration
                // This accounts for nested template rendered elements
                const lastElement = elements[elements.length - 1];
                if (lastElement) {
                    lastInsertedElement = this.findLastRenderedSibling(lastElement);
                }
                
                index++;
            }
        } catch (e) {
            console.error(`Error in p-for refresh: ${iterableExpr}`, e);
        }
    },

    /**
     * Finds the last element rendered by a template (including nested template output)
     */
    findLastRenderedSibling(element) {
        // If this is a p-comp anchor with rendered content, get its last rendered node
        if (element.tagName === 'TEMPLATE' && element._compData && element._compData.renderedElements.length > 0) {
            const lastRendered = element._compData.renderedElements[element._compData.renderedElements.length - 1];
            return this.findLastRenderedSibling(lastRendered);
        }
        // If this is a p-for template with rendered elements, get the last rendered element
        if (element.tagName === 'TEMPLATE' && element._forData && element._forData.renderedElements.length > 0) {
            const lastRendered = element._forData.renderedElements[element._forData.renderedElements.length - 1];
            return this.findLastRenderedSibling(lastRendered);
        }
        return element;
    },

    createLoopScope(parentScope, varPattern, item) {
        const loopData = {};
        varPattern = varPattern.trim();
        
        if (varPattern.startsWith('[')) {
            // Array destructuring: [i, cat]
            const vars = varPattern.slice(1, -1).split(',').map(v => v.trim());
            if (Array.isArray(item)) {
                vars.forEach((v, i) => loopData[v] = item[i]);
            } else {
                loopData[vars[0]] = item;
            }
        } else if (varPattern.startsWith('{')) {
            // Object destructuring: {name, age}
            const vars = varPattern.slice(1, -1).split(',').map(v => v.trim());
            vars.forEach(v => {
                const [key, alias] = v.split(':').map(s => s.trim());
                loopData[alias || key] = item[key];
            });
        } else {
            // Simple variable
            loopData[varPattern] = item;
        }
        
        if (!parentScope) {
            console.error('parentScope is undefined in createLoopScope');
            return new Proxy({}, { get: () => undefined, set: () => false });
        }
        
        const parentTarget = parentScope._p_target || parentScope;
        if (!parentTarget || typeof parentTarget !== 'object') {
            console.error('Invalid parentTarget:', parentTarget);
            return new Proxy(loopData, {
                get: (target, key) => target[key],
                set: (target, key, value) => { target[key] = value; return true; }
            });
        }
        
        const loopTarget = Object.create(parentTarget);
        Object.assign(loopTarget, loopData);
        
        const proxy = new Proxy(loopTarget, {
            get: (target, key) => target[key],
            set: (target, key, value) => {
                if (key in loopData) {
                    target[key] = value;
                } else {
                    parentTarget[key] = value;
                }
                return true;
            }
        });
        proxy._p_target = loopTarget;
        
        return proxy;
    },

    // ==================== MAIN DOM WALKER ====================

    /**
     * Walks the DOM tree, processing scopes, event handlers, and directives
     */
    walkDom(el, parentScope, isHydrating = false) {
        // Skip elements already fully hydrated by hydrateLoop during this hydration pass.
        // This prevents the parent walkDom from re-evaluating inner SSR elements with the
        // wrong (outer-loop) scope when those elements live inside an outer SSR element.
        if (isHydrating && el._p_loopHydrated) {
            return;
        }

        // Handle p-for templates separately
        if (el.tagName === 'TEMPLATE' && el.hasAttribute('p-for')) {
            // Skip nested templates during refresh (non-hydrating) - they will be 
            // re-created by their parent loop. Their scope variables come from the 
            // parent loop iteration, not the static scope chain.
            if (!isHydrating && el._forTemplate) {
                return;
            }
            this.handleFor(el, parentScope, isHydrating);
            return;
        }

        // Handle p-comp templates separately (dynamic components from the registry)
        if (el.tagName === 'TEMPLATE' && el.hasAttribute('p-comp')) {
            // Same ownership rule as nested p-for templates: during refresh the
            // parent loop rebuilds its clones, including their p-comp anchors.
            if (!isHydrating && el._forTemplate) {
                return;
            }
            this.handleComp(el, parentScope, isHydrating);
            return;
        }

        // p-comp on a non-template element is not supported — warn once per element
        if (el.tagName !== 'TEMPLATE' && el.hasAttribute('p-comp') && !el._pCompWarned) {
            el._pCompWarned = true;
            console.warn('p-comp is only supported on <template> elements. Ignoring:', el);
        }

        // Evaluate p-show.pre-scope before determining scope (for Pico conditional components)
        const preScopeShowAttr = Array.from(el.attributes).find(attr => {
            const parsed = this.parseDirectiveModifiers(attr.name);
            return parsed.directive === 'p-show' && parsed.modifiers['pre-scope'];
        });
        if (preScopeShowAttr && parentScope) {
            const value = eval(`with (parentScope) { (${preScopeShowAttr.value}) }`);
            this.directives['p-show'](el, value);
        }

        // Determine the scope for this element
        let currentScope = parentScope;
        
        // Check for p-scope attribute (with or without modifiers like p-scope:sync)
        const hasPScope = Array.from(el.attributes).some(attr => attr.name.startsWith('p-scope'));
        if (hasPScope) {
            if (isHydrating) {
                currentScope = this.initScope(el, parentScope);
            } else {
                currentScope = this.refreshScope(el, parentScope);
            }
        }

        // Store scope on element during hydration
        if (isHydrating) {
            el._scope = currentScope;
            this.registerEventListeners(el);
            this.registerModelBinding(el);
        }
        
        // Evaluate directives (skip pre-scope show as it's already handled)
        if (currentScope) {
            Array.from(el.attributes).forEach(attr => {
                const parsed = this.parseDirectiveModifiers(attr.name);
                // Skip pre-scope show, handle regular p-show and other directives
                if (parsed.directive === 'p-show' && parsed.modifiers['pre-scope']) {
                    return; // Already handled above
                }
                if (Object.keys(this.directives).includes(parsed.directive)) {
                    const evalScope = el._scope || currentScope;
                    const value = eval(`with (evalScope) { (${attr.value}) }`);
                    this.directives[parsed.directive](el, value, parsed.modifiers);
                }
            });
        }

        // During hydration, don't recurse into children of a hidden element.
        // This prevents nested <template p-for> directives from being evaluated with
        // undefined loop variables when their containing p-show element is false
        // (e.g. the inner array loop inside an array-container that is hidden for scalar values).
        if (isHydrating && currentScope) {
            for (const attr of el.attributes) {
                const parsed = this.parseDirectiveModifiers(attr.name);
                if (parsed.directive === 'p-show' && !parsed.modifiers['pre-scope']) {
                    try {
                        if (!eval(`with (currentScope) { (${attr.value}) }`)) {
                            return;
                        }
                    } catch (e) { /* ignore evaluation errors, proceed with children */ }
                    break;
                }
            }
        }

        // Recurse to children
        // Use Array.from to capture a snapshot of children, avoiding issues where
        // p-for templates insert new elements during the walk
        const children = Array.from(el.children);
        for (const child of children) {
            this.walkDom(child, currentScope, isHydrating);
        }
    }


}

window.Pattr.start()
