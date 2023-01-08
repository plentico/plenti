import allContent from '../../generated/content.js';

const validateFilename = (filename, type) => {
    
    let validationErrors = [];

    if (filename.length == 0) {
        validationErrors = [...validationErrors, "Empty filename is not allowed"];
    }
    if (filename.indexOf(' ') >= 0) {
        validationErrors = [...validationErrors, "Spaces not allowed in filename"];
    }
    if (filename.indexOf('~') >= 0) {
        validationErrors = [...validationErrors, "No tilde (~) allowed in filename"];
    }
    if (filename.indexOf('`') >= 0) {
        validationErrors = [...validationErrors, "No backtick (`) allowed in filename"];
    }
    if (filename.indexOf('!') >= 0) {
        validationErrors = [...validationErrors, "No exclamation points (!) allowed in filename"];
    }
    if (filename.indexOf('@') >= 0) {
        validationErrors = [...validationErrors, "No at symbols (@) allowed in filename"];
    }
    if (filename.indexOf('#') >= 0) {
        validationErrors = [...validationErrors, "No pound symbols (#) allowed in filename"];
    }
    if (filename.indexOf('$') >= 0) {
        validationErrors = [...validationErrors, "No dollar signs ($) allowed in filename"];
    }
    if (filename.indexOf('%') >= 0) {
        validationErrors = [...validationErrors, "No percentage symbols (%) allowed in filename"];
    }
    if (filename.indexOf('^') >= 0) {
        validationErrors = [...validationErrors, "No carrot symbol (^) allowed in filename"];
    }
    if (filename.indexOf('&') >= 0) {
        validationErrors = [...validationErrors, "No ampersands (&) allowed in filename"];
    }
    if (filename.indexOf('*') >= 0) {
        validationErrors = [...validationErrors, "No star symbols (*) allowed in filename"];
    }
    if (filename.indexOf('(') >= 0 || filename.indexOf(')') >= 0) {
        validationErrors = [...validationErrors, "No opening or closing round brackets ( ) allowed in filename"];
    }
    if (filename.indexOf('{') >= 0 || filename.indexOf('}') >= 0) {
        validationErrors = [...validationErrors, "No opening or closing curly brackets { } allowed in filename"];
    }
    if (filename.indexOf('[') >= 0 || filename.indexOf(']') >= 0) {
        validationErrors = [...validationErrors, "No opening or closing square brackets [ ] allowed in filename"];
    }
    if (filename.indexOf('<') >= 0 || filename.indexOf('>') >= 0) {
        validationErrors = [...validationErrors, "No opening or closing angle brackets < > allowed in filename"];
    }
    const existingContent = allContent.find(content =>
        content.type == type &&
        content.filename == filename + '.json'
    );
    if (existingContent) {
        validationErrors = [...validationErrors, {message: "Content with this filename already exists.", contentPath: existingContent.path}];
    }

    return validationErrors;
}

export default validateFilename;