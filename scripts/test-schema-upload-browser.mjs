// Real-browser repro of the schema-upload flow against the local plenti serve.
// Run: node scripts/test-schema-upload-browser.mjs
const { chromium } = await import('/home/jim/Projects/plentico/pattr/node_modules/playwright/index.mjs');

const browser = await chromium.launch();
const page = await browser.newPage();
const consoleLogs = [];
page.on('console', msg => consoleLogs.push(msg.type() + ': ' + msg.text()));
page.on('pageerror', err => consoleLogs.push('PAGEERROR: ' + (err.stack || err.message)));

await page.goto('http://localhost:3000/#staff');
await page.click('button.gitlab');           // local login (no OAuth)
await page.waitForTimeout(2000);
await page.screenshot({ path: '/tmp/repro-0-login.png' });
console.log('--- body text after login ---');
console.log((await page.locator('body').innerText()).slice(0, 800));
console.log('--- buttons ---');
for (const b of await page.locator('button').allInnerTexts()) console.log(JSON.stringify(b));

// Open the edit tray for the home page.
const editBtn = page.locator('text=Edit').first();
try {
  await editBtn.waitFor({ timeout: 5000 });
  await editBtn.click();
  await page.waitForTimeout(2500);
} catch { console.log('no Edit button found'); }
await page.screenshot({ path: '/tmp/repro-1-editor.png', fullPage: true });

// Expand cta_image and find its media field.
await page.locator('text=cta_image').first().click();
await page.waitForTimeout(800);
await page.screenshot({ path: '/tmp/repro-2-cta.png' });
// The media field shows a thumbnail; hover reveals "Change Media".
const changeBtn = page.locator('button:has-text("Change Media")');
console.log('--- Change Media buttons:', await changeBtn.count());
if (await changeBtn.count() === 0) {
  // dump visible buttons and inputs in the tray
  console.log('--- buttons:', await page.locator('button').allInnerTexts());
  await browser.close();
  process.exit(1);
}

// ── Upload flow ──────────────────────────────────────────────────────────────
await changeBtn.first().click();
await page.waitForTimeout(1200);
await page.screenshot({ path: '/tmp/repro-3-media-modal.png' });

// Media modal: make sure the Upload tab is active, then set the file input.
const uploadTab = page.locator('button:has-text("Upload")').first();
if (await uploadTab.count()) { try { await uploadTab.click(); } catch {} await page.waitForTimeout(400); }
const fileInput = page.locator('input[type="file"]').first();
await fileInput.setInputFiles('/home/jim/Projects/causeworks/causeworks.com/media/editing.png');
await page.waitForTimeout(2000);
await page.screenshot({ path: '/tmp/repro-4-review-modal.png' });

// The review modal should be open with the schema pre-applied (25x25, webp).
const modalText = await page.locator('body').innerText();
console.log('--- review modal shows 25x25?', /25\s*[x×]\s*25|25\s*\/\s*25/.test(modalText));
console.log('--- review modal shows Reset?', modalText.includes('Reset'));
const dimsMatch = modalText.match(/\d+\s*[x×/]\s*\d+/g);
console.log('--- dims found:', dimsMatch);

// Confirm via the modal's "Next" button (library mode).
await page.locator('button:has-text("Next")').first().click();
await page.waitForTimeout(2000);
await page.screenshot({ path: '/tmp/repro-5-after-confirm.png' });
console.log('--- console after confirm ---');
consoleLogs.forEach(l => console.log(l));

// Save the page and verify content on disk.
await page.locator('button:has-text("Save")').first().click();
await page.waitForTimeout(3000);
console.log('--- console after save ---');
consoleLogs.forEach(l => console.log(l));
await page.screenshot({ path: '/tmp/repro-6-after-save.png' });

console.log('--- console so far ---');
consoleLogs.forEach(l => console.log(l));
await page.screenshot({ path: '/tmp/repro-1-editor.png', fullPage: true });
await browser.close();
