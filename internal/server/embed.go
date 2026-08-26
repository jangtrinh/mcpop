package server

import (
	_ "embed"
	"net/http"
)

// Embedded modern SPA dashboard
const DashboardHTML = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MCPOp — MCP Tool Observability</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      darkMode: 'class',
      theme: {
        extend: {
          colors: {
            brand: {
              50: '#eef2ff',
              500: '#6366f1',
              600: '#4f46e5',
              700: '#4338ca',
            }
          }
        }
      }
    }
  </script>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
  <style>
    @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap');
    body { font-family: 'Plus Jakarta Sans', sans-serif; }
    code, pre, .font-mono { font-family: 'JetBrains Mono', monospace; }
    ::-webkit-scrollbar { width: 6px; height: 6px; }
    ::-webkit-scrollbar-track { background: #0f172a; }
    ::-webkit-scrollbar-thumb { background: #334155; border-radius: 3px; }
  </style>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen flex flex-col antialiased selection:bg-indigo-500 selection:text-white">

  <!-- Top Navigation -->
  <header class="border-b border-slate-800 bg-slate-900/80 backdrop-blur sticky top-0 z-30 px-6 py-3.5 flex items-center justify-between">
    <div class="flex items-center space-x-3">
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20 font-bold text-lg text-white">
        <i class="fa-solid fa-bolt-lightning text-sm"></i>
      </div>
      <div>
        <div class="flex items-center space-x-2">
          <span class="font-bold tracking-tight text-white text-base">MCPOp</span>
          <span class="text-[10px] uppercase font-semibold tracking-wider bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 px-2 py-0.5 rounded-full">Observability</span>
        </div>
        <p class="text-xs text-slate-400">Silent Failure Catcher for MCP Tools</p>
      </div>
    </div>

    <!-- Active Session Selector & Live Pulse -->
    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2 text-xs bg-slate-800/80 px-3 py-1.5 rounded-lg border border-slate-700">
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
        </span>
        <span id="liveStatus" class="font-medium text-emerald-400">Live SSE Connected</span>
      </div>

      <div class="flex items-center space-x-2">
        <select id="sessionSelect" class="bg-slate-800 border border-slate-700 text-xs rounded-lg px-3 py-1.5 text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer">
          <option value="">Loading sessions...</option>
        </select>
        <button onclick="refreshData()" class="p-1.5 hover:bg-slate-800 rounded-lg text-slate-400 hover:text-slate-200 transition">
          <i class="fa-solid fa-arrows-rotate text-xs"></i>
        </button>
      </div>
    </div>
  </header>

  <!-- Main Container -->
  <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-slate-900 border border-slate-800/80 rounded-xl p-4 shadow-sm hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-400 text-xs font-medium">
          <span>Total Tool Calls</span>
          <i class="fa-solid fa-cube text-slate-500"></i>
        </div>
        <div class="text-2xl font-bold text-white mt-1" id="statTotalCalls">0</div>
        <div class="text-[11px] text-slate-500 mt-0.5" id="statCommand">No active command</div>
      </div>

      <div class="bg-slate-900 border border-slate-800/80 rounded-xl p-4 shadow-sm hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-400 text-xs font-medium">
          <span>Success Rate</span>
          <i class="fa-solid fa-shield-check text-slate-500"></i>
        </div>
        <div class="text-2xl font-bold text-emerald-400 mt-1" id="statSuccessRate">100%</div>
        <div class="text-[11px] text-slate-500 mt-0.5" id="statErrorCalls">0 errors</div>
      </div>

      <div class="bg-slate-900 border border-slate-800/80 rounded-xl p-4 shadow-sm hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-400 text-xs font-medium">
          <span>Avg Latency</span>
          <i class="fa-solid fa-stopwatch text-slate-500"></i>
        </div>
        <div class="text-2xl font-bold text-slate-100 mt-1" id="statAvgLatency">0 ms</div>
        <div class="text-[11px] text-slate-500 mt-0.5">Execution speed</div>
      </div>

      <div class="bg-slate-900 border border-slate-800/80 rounded-xl p-4 shadow-sm hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-400 text-xs font-medium">
          <span>Silent Failures Detected</span>
          <i class="fa-solid fa-triangle-exclamation text-amber-500"></i>
        </div>
        <div class="text-2xl font-bold text-rose-400 mt-1" id="statFailures">0</div>
        <div class="text-[11px] text-slate-500 mt-0.5">Loops, Schema, Slow Tools</div>
      </div>
    </div>

    <!-- Failure Alerts Section -->
    <div id="failureAlertsContainer" class="hidden space-y-2">
      <div class="flex items-center space-x-2 text-xs font-semibold uppercase tracking-wider text-rose-400">
        <i class="fa-solid fa-radiation text-rose-500"></i>
        <span>Heuristic Anomaly Detections</span>
      </div>
      <div id="failureAlertsList" class="space-y-2"></div>
    </div>

    <!-- Traces Waterfall Table -->
    <div class="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
      <div class="p-4 border-b border-slate-800 flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <h2 class="text-sm font-bold text-white flex items-center space-x-2">
            <span>Live Tool Trace Waterfall</span>
          </h2>
          <span id="traceCountBadge" class="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full font-mono">0 traces</span>
        </div>
        <div class="flex items-center space-x-2">
          <input id="searchTraces" oninput="renderTraces()" type="text" placeholder="Filter tool name..." class="bg-slate-950 border border-slate-800 text-xs rounded-lg px-3 py-1.5 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 w-48">
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/60 text-slate-400 border-b border-slate-800 font-medium">
            <tr>
              <th class="py-3 px-4">Status</th>
              <th class="py-3 px-4">Tool Name</th>
              <th class="py-3 px-4">Arguments Preview</th>
              <th class="py-3 px-4">Latency</th>
              <th class="py-3 px-4">Timestamp</th>
              <th class="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody id="traceTableBody" class="divide-y divide-slate-800/60 font-mono">
            <tr>
              <td colspan="6" class="py-8 text-center text-slate-500 font-sans">No tool calls recorded in this session yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </main>

  <!-- Tool Detail & Replay Drawer Modal -->
  <div id="detailModal" class="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-50 hidden flex items-center justify-center p-4">
    <div class="bg-slate-900 border border-slate-800 rounded-2xl max-w-3xl w-full shadow-2xl flex flex-col max-h-[90vh]">
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between">
        <div class="flex items-center space-x-2">
          <span class="w-3 h-3 rounded-full bg-indigo-500"></span>
          <h3 class="font-bold text-white text-base font-mono" id="modalToolName">tool/name</h3>
          <span id="modalStatusBadge" class="text-[10px] font-semibold px-2 py-0.5 rounded-full uppercase"></span>
        </div>
        <button onclick="closeModal()" class="text-slate-400 hover:text-white transition">
          <i class="fa-solid fa-xmark text-lg"></i>
        </button>
      </div>

      <div class="p-6 overflow-y-auto space-y-5 text-xs">
        <div>
          <label class="block text-slate-400 font-semibold mb-1 uppercase tracking-wider text-[10px]">Arguments (JSON)</label>
          <textarea id="modalArgsInput" class="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 font-mono text-slate-200 text-xs focus:outline-none focus:border-indigo-500 h-32"></textarea>
        </div>

        <div>
          <label class="block text-slate-400 font-semibold mb-1 uppercase tracking-wider text-[10px]">Execution Result</label>
          <pre id="modalResultText" class="w-full bg-slate-950 border border-slate-800 rounded-xl p-3 font-mono text-emerald-400 text-xs overflow-x-auto max-h-48 whitespace-pre-wrap"></pre>
        </div>

        <!-- Replay Result Box -->
        <div id="replayResultBox" class="hidden p-4 rounded-xl border border-indigo-500/30 bg-indigo-950/20 space-y-2">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-indigo-300">Replay Result</span>
            <span id="replayLatencyBadge" class="text-[11px] font-mono text-indigo-400"></span>
          </div>
          <pre id="replayOutputText" class="font-mono text-xs text-slate-200 whitespace-pre-wrap bg-slate-950 p-2.5 rounded-lg"></pre>
        </div>
      </div>

      <div class="px-6 py-3.5 border-t border-slate-800 bg-slate-950/40 flex items-center justify-between">
        <div class="text-[11px] text-slate-500" id="modalMetadata">Latency: 0ms</div>
        <div class="flex items-center space-x-2">
          <button onclick="closeModal()" class="px-3.5 py-1.5 rounded-lg border border-slate-700 hover:bg-slate-800 text-slate-300 text-xs font-medium transition">Close</button>
          <button id="btnReplay" onclick="runReplay()" class="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition flex items-center space-x-1.5 shadow-lg shadow-indigo-600/30">
            <i class="fa-solid fa-play text-[10px]"></i>
            <span>1-Click Replay</span>
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- JavaScript App Logic -->
  <script>
    let currentSessionId = '';
    let sessions = [];
    let traces = [];
    let failures = [];
    let currentModalTrace = null;

    async function init() {
      await fetchSessions();
      setupSSE();
    }

    async function fetchSessions() {
      try {
        const res = await fetch('/api/sessions');
        sessions = await res.json();
        const select = document.getElementById('sessionSelect');
        select.innerHTML = '';

        if (!sessions || sessions.length === 0) {
          select.innerHTML = '<option value="">No sessions found</option>';
          return;
        }

        sessions.forEach((s, idx) => {
          const opt = document.createElement('option');
          opt.value = s.id;
          opt.textContent = s.command + ' (' + s.id.substring(0, 8) + '...)';
          select.appendChild(opt);
        });

        if (!currentSessionId && sessions.length > 0) {
          currentSessionId = sessions[0].id;
        }
        select.value = currentSessionId;
        select.onchange = (e) => {
          currentSessionId = e.target.value;
          refreshData();
        };

        await refreshData();
      } catch (err) {
        console.error('Failed to fetch sessions:', err);
      }
    }

    async function refreshData() {
      if (!currentSessionId) return;
      try {
        const [tracesRes, failsRes, statsRes] = await Promise.all([
          fetch('/api/sessions/' + currentSessionId + '/traces'),
          fetch('/api/sessions/' + currentSessionId + '/failures'),
          fetch('/api/sessions/' + currentSessionId + '/stats')
        ]);

        traces = await tracesRes.json() || [];
        failures = await failsRes.json() || [];
        const stats = await statsRes.json();

        renderStats(stats);
        renderFailures();
        renderTraces();
      } catch (err) {
        console.error('Failed to refresh data:', err);
      }
    }

    function renderStats(stats) {
      if (!stats) return;
      document.getElementById('statTotalCalls').textContent = stats.total_calls || 0;
      document.getElementById('statSuccessRate').textContent = (stats.success_rate || 100).toFixed(1) + '%';
      document.getElementById('statErrorCalls').textContent = (stats.error_calls || 0) + ' errors';
      document.getElementById('statAvgLatency').textContent = (stats.avg_latency_ms || 0) + ' ms';
      document.getElementById('statFailures').textContent = stats.failure_count || 0;

      const curr = sessions.find(s => s.id === currentSessionId);
      if (curr) {
        document.getElementById('statCommand').textContent = curr.command;
      }
    }

    function renderFailures() {
      const container = document.getElementById('failureAlertsContainer');
      const list = document.getElementById('failureAlertsList');
      if (!failures || failures.length === 0) {
        container.classList.add('hidden');
        return;
      }
      container.classList.remove('hidden');
      list.innerHTML = '';

      failures.forEach(f => {
        const div = document.createElement('div');
        let badgeColor = 'bg-rose-500/10 text-rose-400 border-rose-500/20';
        let icon = 'fa-repeat';
        if (f.failure_type === 'schema_mismatch') {
          badgeColor = 'bg-amber-500/10 text-amber-400 border-amber-500/20';
          icon = 'fa-file-code';
        } else if (f.failure_type === 'slow_tool') {
          badgeColor = 'bg-orange-500/10 text-orange-400 border-orange-500/20';
          icon = 'fa-hourglass-half';
        }

        div.className = 'p-3 rounded-xl border flex items-start space-x-3 ' + badgeColor;
        div.innerHTML = '<i class="fa-solid ' + icon + ' mt-0.5 text-sm"></i>' +
          '<div class="flex-1">' +
            '<div class="flex items-center justify-between">' +
              '<span class="font-bold text-xs uppercase tracking-wider">' + f.failure_type.replace('_', ' ') + '</span>' +
              '<span class="text-[10px] opacity-70">' + new Date(f.created_at).toLocaleTimeString() + '</span>' +
            '</div>' +
            '<p class="text-xs mt-0.5 text-slate-200">' + f.description + '</p>' +
          '</div>';
        list.appendChild(div);
      });
    }

    function renderTraces() {
      const tbody = document.getElementById('traceTableBody');
      const search = document.getElementById('searchTraces').value.toLowerCase();
      document.getElementById('traceCountBadge').textContent = traces.length + ' traces';

      const filtered = traces.filter(t => t.tool_name.toLowerCase().includes(search));

      if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="py-8 text-center text-slate-500 font-sans">No tool calls found.</td></tr>';
        return;
      }

      tbody.innerHTML = '';
      filtered.forEach(t => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-slate-800/40 transition cursor-pointer';

        let statusBadge = '<span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">SUCCESS</span>';
        if (t.is_error || t.status === 'failed') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20">FAILED</span>';
        } else if (t.status === 'pending') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 animate-pulse">RUNNING</span>';
        }

        let latColor = 'text-emerald-400';
        if (t.latency_ms > 1000) latColor = 'text-amber-400';
        if (t.latency_ms >= 5000) latColor = 'text-rose-400 font-bold';

        let argsPreview = t.arguments;
        if (argsPreview && argsPreview.length > 50) {
          argsPreview = argsPreview.substring(0, 50) + '...';
        }

        tr.innerHTML = 
          '<td class="py-3 px-4">' + statusBadge + '</td>' +
          '<td class="py-3 px-4 font-bold text-slate-200">' + t.tool_name + '</td>' +
          '<td class="py-3 px-4 text-slate-400 truncate max-w-xs">' + argsPreview + '</td>' +
          '<td class="py-3 px-4 ' + latColor + '">' + t.latency_ms + ' ms</td>' +
          '<td class="py-3 px-4 text-slate-500 text-[11px]">' + new Date(t.created_at).toLocaleTimeString() + '</td>' +
          '<td class="py-3 px-4 text-right space-x-2 font-sans">' +
            '<button onclick="inspectTrace(\'' + t.id + '\')" class="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px] transition">Inspect</button>' +
          '</td>';

        tr.onclick = (e) => {
          if (e.target.tagName !== 'BUTTON') {
            inspectTrace(t.id);
          }
        };
        tbody.appendChild(tr);
      });
    }

    function inspectTrace(id) {
      const trace = traces.find(t => t.id === id);
      if (!trace) return;
      currentModalTrace = trace;

      document.getElementById('modalToolName').textContent = trace.tool_name;
      const statusBadge = document.getElementById('modalStatusBadge');
      if (trace.is_error) {
        statusBadge.textContent = 'Failed';
        statusBadge.className = 'text-[10px] font-semibold px-2 py-0.5 rounded-full uppercase bg-rose-500/10 text-rose-400 border border-rose-500/20';
      } else {
        statusBadge.textContent = 'Success';
        statusBadge.className = 'text-[10px] font-semibold px-2 py-0.5 rounded-full uppercase bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
      }

      try {
        const parsed = JSON.parse(trace.arguments);
        document.getElementById('modalArgsInput').value = JSON.stringify(parsed, null, 2);
      } catch {
        document.getElementById('modalArgsInput').value = trace.arguments;
      }

      document.getElementById('modalResultText').textContent = trace.result || trace.error_message || '(No result returned)';
      document.getElementById('modalMetadata').textContent = 'Latency: ' + trace.latency_ms + ' ms | ID: ' + trace.rpc_id;
      document.getElementById('replayResultBox').classList.add('hidden');
      document.getElementById('detailModal').classList.remove('hidden');
    }

    function closeModal() {
      document.getElementById('detailModal').classList.add('hidden');
    }

    async function runReplay() {
      if (!currentModalTrace) return;
      const btn = document.getElementById('btnReplay');
      const box = document.getElementById('replayResultBox');
      const out = document.getElementById('replayOutputText');
      const lat = document.getElementById('replayLatencyBadge');

      btn.disabled = true;
      btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin text-xs"></i> <span>Replaying...</span>';

      const session = sessions.find(s => s.id === currentSessionId);
      const command = session ? session.command : '';

      let parsedArgs = {};
      try {
        parsedArgs = JSON.parse(document.getElementById('modalArgsInput').value);
      } catch (e) {
        alert('Invalid JSON in arguments');
        btn.disabled = false;
        btn.innerHTML = '<i class="fa-solid fa-play text-[10px]"></i> <span>1-Click Replay</span>';
        return;
      }

      try {
        const res = await fetch('/api/replay', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            command: command,
            tool_name: currentModalTrace.tool_name,
            arguments: parsedArgs
          })
        });

        const data = await res.json();
        box.classList.remove('hidden');
        lat.textContent = 'Latency: ' + data.latency_ms + 'ms';
        out.textContent = JSON.stringify(data.result || data.error, null, 2);
      } catch (err) {
        box.classList.remove('hidden');
        lat.textContent = 'Error';
        out.textContent = 'Replay failed: ' + err.message;
      } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="fa-solid fa-play text-[10px]"></i> <span>1-Click Replay</span>';
      }
    }

    function setupSSE() {
      const evtSource = new EventSource('/api/events');
      evtSource.onopen = () => {
        document.getElementById('liveStatus').textContent = 'Live SSE Connected';
      };
      evtSource.onerror = () => {
        document.getElementById('liveStatus').textContent = 'Reconnecting...';
      };
      evtSource.onmessage = (e) => {
        try {
          refreshData();
        } catch {}
      };
      evtSource.addEventListener('tool_call', (e) => {
        refreshData();
      });
      evtSource.addEventListener('failure', (e) => {
        refreshData();
      });
    }

    window.onload = init;
  </script>
</body>
</html>`

// Handler serves embedded dashboard
func DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(DashboardHTML))
	}
}
