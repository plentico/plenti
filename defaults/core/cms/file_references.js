import allContent from '../../generated/content.js';

const checkType = (value, file) => {
	if (typeof value === 'string') {
  	    if (value.includes(file)) {
    	    return true;
   	    }
 	}
    if (value && value.constructor === ({}).constructor) {
        // It's an Obj, so go another layer deep
        return findFile(value, file);
    }
    return false;
}

const findFile = (fields, file) => {
    let found = false;
    for (const [key, value] of Object.entries(fields)) {
        if (value && value.constructor === [].constructor) {
            // It's an array, so loop through values
    	    value.forEach(nestedVal => {
                // Check the type of each value
				if (checkType(nestedVal, file)) {
        	        found = true;
                };
            });
        } else {
    	    if (checkType(value, file)) {
      	        found = true;
            };
        }
    }
    return found;
}

export const findFileReferences = file => {
    let found = [];
    allContent.forEach(content => {
	    if (findFile(content.fields, file)) {
  	        found = [...found, content.path];
        }
    });
    return found;
}