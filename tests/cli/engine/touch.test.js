/**
 * CLI Tests: touch gestures (tap, swipe, long-press).
 * Each command goes daemon -> agent handler -> api touch primitive, the same
 * path browser_tap/swipe/long_press take over MCP (#682). Gated by the touch
 * capability since Firefox desktop ships touch events off.
 */

const { test, describe } = require("../../helpers/capabilities").suite("touch");
const assert = require('node:assert');
const { execSync } = require('node:child_process');
const { VIBIUM } = require("../../helpers");

function sh(cmd) {
  return execSync(`${VIBIUM} ${cmd}`, { encoding: 'utf-8', timeout: 30000 });
}

// Record pointer events at the document level (capture phase) so the gesture
// is observed wherever the daemon's current page points. touch-action:none
// stops the browser turning a swipe into a scroll (which fires pointercancel
// partway). Reset __ev each call; install the listeners once. Double quotes are
// avoided so the whole expression can be wrapped in shell double quotes.
const RECORD = "eval \"(function(){document.documentElement.style.touchAction='none';document.body.style.touchAction='none';window.__ev=[];if(!window.__touchrec){window.__touchrec=true;for(const t of ['pointerdown','pointermove','pointerup'])document.addEventListener(t,e=>window.__ev.push({type:t,x:e.clientX,y:e.clientY,pt:e.pointerType}),true);}return 'ok'})()\"";

describe('CLI: touch gestures', () => {
  test('tap confirms and drives a touch pointerdown at the point', () => {
    sh(RECORD);
    assert.match(sh(`tap 60 80`), /Tapped at \(60, 80\)/);
    const summary = sh(`eval "(function(){var d=window.__ev.find(e=>e.type==='pointerdown');return [d.pt,d.x,d.y].join('|')})()"`);
    assert.match(summary, /touch\|60\|80/, 'tap records a touch pointerdown at (60, 80)');
  });

  test('swipe presses at start, moves while held, releases at end', () => {
    sh(RECORD);
    assert.match(sh(`swipe 40 100 240 100 --duration 200`), /Swiped from \(40, 100\) to \(240, 100\)/);
    const summary = sh(`eval "(function(){var e=window.__ev;return [e.every(x=>x.pt==='touch'),e.find(x=>x.type==='pointerdown').x,e.filter(x=>x.type==='pointermove').length>0,e[e.length-1].type,e[e.length-1].x].join('|')})()"`);
    assert.match(summary, /true\|40\|true\|pointerup\|240/, 'swipe drags a held touch pointer from start to end');
  });

  test('long-press holds a touch pointer then releases', () => {
    sh(RECORD);
    assert.match(sh(`long-press 120 120 --duration 300`), /Long-pressed at \(120, 120\)/);
    const summary = sh(`eval "(function(){var e=window.__ev;var d=e.find(x=>x.type==='pointerdown');return [d.pt,d.x,d.y,!!e.find(x=>x.type==='pointerup')].join('|')})()"`);
    assert.match(summary, /touch\|120\|120\|true/, 'long press holds then releases a touch pointer');
  });
});
