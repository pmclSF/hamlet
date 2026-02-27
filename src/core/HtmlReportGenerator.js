import fs from 'fs/promises';
import path from 'path';

/**
 * Generate a static dark-themed HTML report from an analysis report object.
 *
 * Outputs:
 *   <outDir>/index.html   — self-contained, works offline
 *   <outDir>/report.json  — raw JSON sidecar
 *
 * @param {Object} report - Analysis report from ProjectAnalyzer
 * @param {string} outDir - Directory to write into (created if missing)
 */
export async function generateHtmlReport(report, outDir) {
  const resolved = path.resolve(outDir);
  await fs.mkdir(resolved, { recursive: true });

  // Write JSON sidecar
  await fs.writeFile(
    path.join(resolved, 'report.json'),
    JSON.stringify(report, null, 2)
  );

  // Write HTML
  const html = buildHtml(report);
  await fs.writeFile(path.join(resolved, 'index.html'), html);
}

function esc(s) {
  if (s == null) return '';
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function badge(text, variant) {
  return `<span class="badge badge-${variant || 'default'}">${esc(text)}</span>`;
}

function frameworkBadge(fw) {
  if (!fw) return badge('unknown', 'muted');
  const colors = {
    jest: 'success',
    vitest: 'success',
    mocha: 'warning',
    jasmine: 'warning',
    cypress: 'accent',
    playwright: 'accent',
    selenium: 'muted',
    pytest: 'info',
    unittest: 'info',
    nose2: 'info',
    junit4: 'danger',
    junit5: 'danger',
    testng: 'danger',
    webdriverio: 'muted',
    puppeteer: 'muted',
    testcafe: 'muted',
  };
  return badge(fw, colors[fw] || 'default');
}

function confidenceHtml(val) {
  const pct = Math.round(val);
  let variant = 'danger';
  if (pct >= 80) variant = 'success';
  else if (pct >= 50) variant = 'warning';
  return `<span class="confidence">${badge(pct + '%', variant)}<span class="bar"><span class="bar-fill bar-${variant}" style="width:${pct}%"></span></span></span>`;
}

function buildFileRows(files) {
  return files
    .map(
      (f, i) =>
        `<tr class="file-row" data-idx="${i}">
      <td class="mono">${esc(f.path)}</td>
      <td>${badge(f.type, f.type === 'test' ? 'success' : 'muted')}</td>
      <td>${frameworkBadge(f.framework)}</td>
      <td data-val="${f.confidence}">${confidenceHtml(f.confidence)}</td>
      <td>${
        f.candidates.length > 0
          ? f.candidates
              .slice(0, 2)
              .map((c) => badge(c.framework, 'muted'))
              .join(' ')
          : '--'
      }</td>
      <td>${f.warnings.length > 0 ? badge(f.warnings.length, 'warning') : '--'}</td>
    </tr>`
    )
    .join('\n');
}

function buildDirections(dirs) {
  if (!dirs || dirs.length === 0) return '';
  return `<div class="directions"><h3>Supported Directions</h3><div class="dir-list">${dirs
    .map(
      (d) =>
        `<div class="dir-item">${frameworkBadge(d.from)} <span class="dir-arrow">\u2192</span> ${frameworkBadge(d.to)} ${d.pipelineBacked ? badge('pipeline', 'success') : badge('legacy', 'muted')}</div>`
    )
    .join('\n')}</div></div>`;
}

function buildHtml(report) {
  const { meta, summary, files } = report;
  const jsonBlob = JSON.stringify(report);

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Hamlet Analysis Report</title>
<style>
:root{--bg-0:#0d1117;--bg-1:#161b22;--bg-2:#21262d;--bg-3:#30363d;--text-0:#e6edf3;--text-1:#8b949e;--text-2:#484f58;--accent:#58a6ff;--accent-dim:#1f6feb;--success:#3fb950;--warning:#d29922;--danger:#f85149;--info:#79c0ff;--radius:6px;--mono:'SF Mono','Cascadia Code','Fira Code',Consolas,monospace;--sans:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%;background:var(--bg-0);color:var(--text-0);font-family:var(--sans);font-size:14px;line-height:1.5;-webkit-font-smoothing:antialiased}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}
code,.mono{font-family:var(--mono);font-size:13px}
.topbar{position:sticky;top:0;z-index:100;display:flex;align-items:center;justify-content:space-between;padding:0 24px;height:48px;background:var(--bg-1);border-bottom:1px solid var(--bg-3)}
.topbar-brand{font-weight:600;font-size:16px}
.topbar-actions{display:flex;gap:8px}
.container{max-width:1400px;margin:0 auto;padding:24px}
.meta{font-size:12px;color:var(--text-2);margin-bottom:16px}
.meta span{margin-right:16px}
.cards-row{display:flex;gap:12px;margin-bottom:20px;flex-wrap:wrap}
.card{flex:1;min-width:140px;padding:16px;background:var(--bg-1);border:1px solid var(--bg-3);border-radius:var(--radius)}
.card-value{font-size:24px;font-weight:600;line-height:1.2}
.card-label{font-size:12px;color:var(--text-1);margin-top:4px;text-transform:uppercase;letter-spacing:.05em}
.card-sub{font-size:12px;color:var(--text-2);margin-top:4px}
.directions{margin-bottom:20px}
.directions h3{font-size:14px;margin-bottom:8px}
.dir-list{display:flex;flex-wrap:wrap;gap:8px}
.dir-item{display:inline-flex;align-items:center;gap:4px;padding:4px 10px;background:var(--bg-1);border:1px solid var(--bg-3);border-radius:var(--radius);font-size:13px}
.dir-arrow{color:var(--text-2)}
.toolbar{display:flex;gap:8px;margin-bottom:12px;align-items:center}
.search-input{padding:6px 10px;background:var(--bg-1);border:1px solid var(--bg-3);border-radius:var(--radius);color:var(--text-0);font-size:13px;font-family:var(--sans);outline:none;width:260px}
.search-input:focus{border-color:var(--accent)}
.split{display:flex;gap:16px}
.table-side{flex:1;min-width:0}
.detail-side{width:0;overflow:hidden;transition:width .2s;flex-shrink:0}
.detail-side.open{width:380px;border-left:1px solid var(--bg-3);padding-left:16px}
.table-wrap{overflow-x:auto;border:1px solid var(--bg-3);border-radius:var(--radius)}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:8px 12px;background:var(--bg-1);border-bottom:1px solid var(--bg-3);color:var(--text-1);font-weight:500;font-size:12px;text-transform:uppercase;letter-spacing:.04em;white-space:nowrap;user-select:none;cursor:pointer}
th:hover{color:var(--text-0)}
td{padding:6px 12px;border-bottom:1px solid var(--bg-2);vertical-align:middle}
tbody tr{transition:background .1s}
tbody tr:hover{background:var(--bg-1)}
tbody tr.selected{background:var(--bg-2)}
.badge{display:inline-block;padding:2px 8px;border-radius:12px;font-size:11px;font-weight:500;line-height:1.4;white-space:nowrap}
.badge-default{background:var(--bg-2);color:var(--text-1)}.badge-muted{background:var(--bg-2);color:var(--text-1)}
.badge-success{background:rgba(63,185,80,.15);color:var(--success)}.badge-warning{background:rgba(210,153,34,.15);color:var(--warning)}
.badge-danger{background:rgba(248,81,73,.15);color:var(--danger)}.badge-accent{background:rgba(88,166,255,.15);color:var(--accent)}
.badge-info{background:rgba(121,192,255,.15);color:var(--info)}
.confidence{display:inline-flex;align-items:center;gap:6px}
.bar{width:60px;height:4px;background:var(--bg-2);border-radius:2px;overflow:hidden;display:inline-block}
.bar-fill{height:100%;border-radius:2px;display:block}.bar-success{background:var(--success)}.bar-warning{background:var(--warning)}.bar-danger{background:var(--danger)}
.btn{display:inline-flex;align-items:center;gap:6px;padding:6px 14px;border:1px solid var(--bg-3);border-radius:var(--radius);background:var(--bg-1);color:var(--text-0);font-size:13px;font-family:var(--sans);font-weight:500;cursor:pointer;white-space:nowrap}
.btn:hover{background:var(--bg-2)}.btn:focus-visible{outline:2px solid var(--accent);outline-offset:1px}
.btn-ghost{background:transparent;border-color:transparent}.btn-ghost:hover{background:var(--bg-2)}
.detail-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}
.detail-header h3{font-size:14px;font-family:var(--mono);word-break:break-all}
.detail-section{margin-bottom:16px;padding-bottom:16px;border-bottom:1px solid var(--bg-2)}
.detail-section h4{font-size:12px;color:var(--text-1);text-transform:uppercase;letter-spacing:.04em;margin-bottom:8px}
.detail-row{display:flex;align-items:center;gap:8px;margin-bottom:6px;font-size:13px}
.detail-label{width:80px;flex-shrink:0;color:var(--text-1);font-size:12px}
.detail-empty{display:flex;align-items:center;justify-content:center;height:200px;color:var(--text-2);font-size:13px}
.command-block{display:flex;align-items:center;gap:8px;padding:8px 12px;background:var(--bg-0);border:1px solid var(--bg-3);border-radius:var(--radius);font-family:var(--mono);font-size:12px;word-break:break-all}
.command-block code{flex:1}
.hidden{display:none!important}
::-webkit-scrollbar{width:8px;height:8px}::-webkit-scrollbar-track{background:transparent}::-webkit-scrollbar-thumb{background:var(--bg-3);border-radius:4px}::-webkit-scrollbar-thumb:hover{background:var(--text-2)}
</style>
</head>
<body>
<header class="topbar">
  <div class="topbar-brand">Hamlet Analysis Report</div>
  <div class="topbar-actions">
    <button class="btn" id="dl-json">Download JSON</button>
  </div>
</header>
<div class="container">
  <div class="meta">
    <span>Hamlet v${esc(meta.hamletVersion)}</span>
    <span>Generated ${esc(meta.generatedAt)}</span>
    <span>Root: <code>${esc(meta.root)}</code></span>
  </div>
  <div class="cards-row">
    <div class="card"><div class="card-value">${summary.fileCount}</div><div class="card-label">Files Scanned</div></div>
    <div class="card"><div class="card-value">${summary.testFileCount}</div><div class="card-label">Test Files</div></div>
    <div class="card"><div class="card-value">${summary.frameworksDetected.length}</div><div class="card-label">Frameworks</div><div class="card-sub">${esc(summary.frameworksDetected.join(', ') || 'none')}</div></div>
    <div class="card"><div class="card-value">${summary.confidenceAvg > 0 ? summary.confidenceAvg + '%' : '--'}</div><div class="card-label">Avg Confidence</div></div>
    <div class="card"><div class="card-value">${summary.directionsSupported.length}</div><div class="card-label">Directions</div></div>
  </div>
  ${buildDirections(summary.directionsSupported)}
  <div class="toolbar">
    <input type="text" class="search-input" id="search" placeholder="Search files\u2026" />
  </div>
  <div class="split">
    <div class="table-side">
      <div class="table-wrap">
        <table id="file-table">
          <thead><tr>
            <th data-col="path">Path</th>
            <th data-col="type">Type</th>
            <th data-col="framework">Framework</th>
            <th data-col="confidence">Confidence</th>
            <th>Candidates</th>
            <th>Warnings</th>
          </tr></thead>
          <tbody id="tbody">${buildFileRows(files)}</tbody>
        </table>
      </div>
    </div>
    <div class="detail-side" id="detail"><div class="detail-empty">Select a file to view details</div></div>
  </div>
</div>
<script>
(function(){
var DATA=${jsonBlob};
var files=DATA.files;
var dirs=DATA.summary.directionsSupported;
var sortCol='path',sortAsc=true;
var tbody=document.getElementById('tbody');
var detail=document.getElementById('detail');
var search=document.getElementById('search');

function esc(s){return s==null?'':String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}
function badge(t,v){return '<span class="badge badge-'+(v||'default')+'">'+esc(t)+'</span>'}
var fwc={jest:'success',vitest:'success',mocha:'warning',jasmine:'warning',cypress:'accent',playwright:'accent',selenium:'muted',pytest:'info',unittest:'info',nose2:'info',junit4:'danger',junit5:'danger',testng:'danger',webdriverio:'muted',puppeteer:'muted',testcafe:'muted'};
function fwBadge(fw){return fw?badge(fw,fwc[fw]||'default'):badge('unknown','muted')}
function confHtml(v){var p=Math.round(v),c=p>=80?'success':p>=50?'warning':'danger';return '<span class="confidence">'+badge(p+'%',c)+'<span class="bar"><span class="bar-fill bar-'+c+'" style="width:'+p+'%"></span></span></span>'}

function renderRows(list){
  tbody.innerHTML=list.map(function(f,i){
    return '<tr class="file-row" data-idx="'+f._idx+'"><td class="mono">'+esc(f.path)+'</td><td>'+badge(f.type,f.type==='test'?'success':'muted')+'</td><td>'+fwBadge(f.framework)+'</td><td data-val="'+f.confidence+'">'+confHtml(f.confidence)+'</td><td>'+(f.candidates.length>0?f.candidates.slice(0,2).map(function(c){return badge(c.framework,'muted')}).join(' '):'--')+'</td><td>'+(f.warnings.length>0?badge(f.warnings.length,'warning'):'--')+'</td></tr>';
  }).join('');
}

function getFiltered(){
  var q=(search.value||'').toLowerCase();
  var list=files.map(function(f,i){f._idx=i;return f});
  if(q)list=list.filter(function(f){return f.path.toLowerCase().indexOf(q)>=0});
  list.sort(function(a,b){
    var va=a[sortCol],vb=b[sortCol];
    if(va==null)va='';if(vb==null)vb='';
    if(typeof va==='number'&&typeof vb==='number')return sortAsc?va-vb:vb-va;
    va=String(va).toLowerCase();vb=String(vb).toLowerCase();
    return sortAsc?(va<vb?-1:va>vb?1:0):(va>vb?-1:va<vb?1:0);
  });
  return list;
}

function refresh(){renderRows(getFiltered())}

document.querySelectorAll('th[data-col]').forEach(function(th){
  th.addEventListener('click',function(){
    var col=th.getAttribute('data-col');
    if(sortCol===col)sortAsc=!sortAsc;else{sortCol=col;sortAsc=true}
    document.querySelectorAll('th[data-col]').forEach(function(t){t.textContent=t.textContent.replace(/ [\\u25B2\\u25BC]/,'')});
    th.textContent=th.textContent+(sortAsc?' \\u25B2':' \\u25BC');
    refresh();
  });
});

search.addEventListener('input',refresh);

tbody.addEventListener('click',function(e){
  var row=e.target.closest('.file-row');
  if(!row)return;
  var idx=parseInt(row.getAttribute('data-idx'));
  var f=files[idx];if(!f)return;
  tbody.querySelectorAll('tr').forEach(function(r){r.classList.remove('selected')});
  row.classList.add('selected');
  var fd=dirs.filter(function(d){return d.from===f.framework});
  var cmd=fd.length>0?'hamlet convert '+f.path+' --from '+f.framework+' --to '+fd[0].to+' -o hamlet-out/':'';
  detail.className='detail-side open';
  detail.innerHTML='<div class="detail-header"><h3>'+esc(f.path)+'</h3><button class="btn btn-ghost" id="close-detail">\\u2715</button></div>'
    +'<div class="detail-section"><div class="detail-row"><span class="detail-label">Type</span>'+badge(f.type,f.type==='test'?'success':'muted')+'</div>'
    +'<div class="detail-row"><span class="detail-label">Framework</span>'+fwBadge(f.framework)+'</div>'
    +'<div class="detail-row"><span class="detail-label">Confidence</span>'+confHtml(f.confidence)+'</div></div>'
    +(f.candidates.length>0?'<div class="detail-section"><h4>Detection Candidates</h4>'+f.candidates.map(function(c){return '<div class="detail-row">'+fwBadge(c.framework)+' <span style="color:var(--text-1)">score: '+c.score+'</span></div>'}).join('')+'</div>':'')
    +(f.warnings.length>0?'<div class="detail-section"><h4>Warnings</h4>'+f.warnings.map(function(w){return '<div style="padding:4px 8px;margin-bottom:4px;background:rgba(210,153,34,.1);border-left:3px solid var(--warning);border-radius:0 var(--radius) var(--radius) 0;font-size:12px;color:var(--warning)">'+esc(w)+'</div>'}).join('')+'</div>':'')
    +(cmd?'<div class="detail-section"><h4>Recommended Command</h4><div class="command-block"><code>'+esc(cmd)+'</code><button class="btn btn-ghost" style="padding:3px 6px;font-size:11px" onclick="navigator.clipboard.writeText(this.previousElementSibling.textContent);this.textContent=\\'Copied!\\';setTimeout(function(){}.bind(this),1500)">Copy</button></div></div>':'')
    +(fd.length>0?'<div class="detail-section"><h4>Available Targets</h4>'+fd.map(function(d){return '<div class="detail-row">'+fwBadge(d.to)+' '+(d.pipelineBacked?badge('pipeline','success'):badge('legacy','muted'))+'</div>'}).join('')+'</div>':'');
  document.getElementById('close-detail').addEventListener('click',function(){
    detail.className='detail-side';detail.innerHTML='<div class="detail-empty">Select a file to view details</div>';
    tbody.querySelectorAll('tr').forEach(function(r){r.classList.remove('selected')});
  });
});

document.getElementById('dl-json').addEventListener('click',function(){
  var blob=new Blob([JSON.stringify(DATA,null,2)],{type:'application/json'});
  var a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download='hamlet-analysis.json';a.click();URL.revokeObjectURL(a.href);
});
})();
</script>
</body>
</html>`;
}
