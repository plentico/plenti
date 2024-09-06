<script context="module">
	
	//JavaScript HTML Sanitizer v2.0.2, (c) Alexander Yumashev, Jitbit Software.
	//homepage https://github.com/jitbit/HtmlSanitizer
	//License: MIT https://github.com/jitbit/HtmlSanitizer/blob/master/LICENSE
	//Modified to fit project needs
	
	// https://developer.mozilla.org/en-US/docs/Web/API/Element/paste_event
	export const plaintextPaste = e => {
		e.preventDefault();
		const data = e.clipboardData || window.clipboardData;
		const raw = data.getData("text/html") || data.getData("text");
		const paste = HtmlSanitizer.SanitizeHtml(raw);
		document.execCommand("insertHTML", false, paste);
	}
	
	const HtmlSanitizer = typeof window === "undefined" ? null : new (function () {
	
		const _tagWhitelist = {
			'A':          true,
			'ABBR':       true,
			'B':          true,
			'BLOCKQUOTE': true,
			'BODY':       true,
			'BR':         true,
			'CENTER':     true,
			'CODE':       true,
			'DD':         true,
			'DIV':        true,
			'DL':         true,
			'DT':         true,
			'EM':         true,
			'FONT':       true,
			'H1':         true,
			'H2':         true,
			'H3':         true,
			'H4':         true,
			'H5':         true,
			'H6':         true,
			'HR':         true,
			'I':          true,
			'IMG':        true,
			'LABEL':      true,
			'LI':         true,
			'OL':         true,
			'P':          true,
			'PRE':        true,
			'S':          true,
			'STRIKE':     true,
			'SMALL':      true,
			'SOURCE':     true,
			'SPAN':       true,
			'STRONG':     true,
			'SUB':        true,
			'SUP':        true,
			'TABLE':      true,
			'TBODY':      true,
			'TR':         true,
			'TD':         true,
			'TH':         true,
			'THEAD':      true,
			'UL':         true,
			'U':          true,
			'VIDEO':      true,
		};

		// tags that will be converted to new mapped tag
		// note: must be in whitelist
		const _tagConvertList = {
			'STRIKE': 'S',
		};
	
		// tags that will be converted to DIVs
		const _contentTagWhiteList = {
			'FORM':                      true,
			'GOOGLE-SHEETS-HTML-ORIGIN': true,
			'P':                         true,
		};
	
		const _attributeWhitelist = {
			'align':    true,
			'color':    true,
			'controls': true,
			'height':   true,
			'href':     true,
			'id':       true,
			'src':      true,
			'style':    false,
			'target':   true,
			'title':    true,
			'type':     true,
			'width':    true,
		};
	
		const _cssWhitelist = {
			'background-color': false,
			'color':            false,
			'font-size':        false,
			'font-weight':      false,
			'text-align':       false,
			'text-decoration':  false,
			'width':            false,
		};
	
		// which "protocols" are allowed in "href", "src" etc
		const _schemaWhiteList = [
			'http:',
			'https:',
			'data:',
			'm-files:',
			'file:',
			'ftp:',
			'mailto:',
			'pw:',
		];
	
		const _uriAttributes = {
			'href': true,
			'action': true,
		};
	
		const _parser = new DOMParser();
	
		this.SanitizeHtml = function (input, extraSelector) {
			input = input.trim();
			if (input == "") return ""; //to save performance
	
			//firefox "bogus node" workaround for wysiwyg's
			if (input == "<br>") return "";
	
			if (input.indexOf("<body")==-1) input = "<body>" + input + "</body>"; //add "body" otherwise some tags are skipped, like <style>
	
			let doc = _parser.parseFromString(input, "text/html");
	
			//DOM clobbering check (damn you firefox)
			if (doc.body.tagName !== 'BODY')
				doc.body.remove();
			if (typeof doc.createElement !== 'function')
				doc.createElement.remove();
	
			function makeSanitizedCopy(node) {
				let newNode;
				if (node.nodeType == Node.TEXT_NODE) {
					newNode = node.cloneNode(true);
				} else if (node.nodeType == Node.ELEMENT_NODE && (_tagWhitelist[node.tagName] || _contentTagWhiteList[node.tagName] || (extraSelector && node.matches(extraSelector)))) { //is tag allowed?
	
					if (_contentTagWhiteList[node.tagName])
						newNode = doc.createElement('DIV'); //convert to DIV
					else if (_tagConvertList[node.tagName])
						newNode = doc.createElement(_tagConvertList[node.tagName]); //convert to mapped tag
					else
						newNode = doc.createElement(node.tagName);
	
					for (let i = 0; i < node.attributes.length; i++) {
						let attr = node.attributes[i];
						if (_attributeWhitelist[attr.name]) {
							if (attr.name == "style") {
								for (let s = 0; s < node.style.length; s++) {
									let styleName = node.style[s];
									if (_cssWhitelist[styleName])
										newNode.style.setProperty(styleName, node.style.getPropertyValue(styleName));
								}
							}
							else {
								if (_uriAttributes[attr.name]) { //if this is a "uri" attribute, that can have "javascript:" or something
									if (attr.value.indexOf(":") > -1 && !startsWithAny(attr.value, _schemaWhiteList))
										continue;
								}
								newNode.setAttribute(attr.name, attr.value);
							}
						}
					}
					for (let i = 0; i < node.childNodes.length; i++) {
						let subCopy = makeSanitizedCopy(node.childNodes[i]);
						newNode.appendChild(subCopy, false);
					}
	
					//remove useless empty spans (lots of those when pasting from MS Outlook)
					if ((newNode.tagName == "SPAN" || newNode.tagName == "B" || newNode.tagName == "I" || newNode.tagName == "U")
							&& newNode.innerHTML.trim() == "") {
						return doc.createDocumentFragment();
					}
	
				} else {
					newNode = doc.createDocumentFragment();
				}
				return newNode;
			};
	
			let resultElement = makeSanitizedCopy(doc.body);
	
			return resultElement.innerHTML
				.replace(/<br[^>]*>(\S)/g, "<br>\n$1")
				.replace(/div><div/g, "div>\n<div"); //replace is just for cleaner code
		}
	
		function startsWithAny(str, substrings) {
			for (let i = 0; i < substrings.length; i++) {
				if (str.indexOf(substrings[i]) == 0) {
					return true;
				}
			}
			return false;
		}
	
		this.AllowedTags = _tagWhitelist;
		this.AllowedAttributes = _attributeWhitelist;
		this.AllowedCssStyles = _cssWhitelist;
		this.AllowedSchemas = _schemaWhiteList;
	});
</script>
