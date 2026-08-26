package server

import (
	_ "embed"
	"net/http"
)

// Embedded sleek Neutral B&W Dashboard with Functional Semantic Signals (Green/Red/Amber)
const DashboardHTML = `<!DOCTYPE html>
<html lang="en" class="light">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MCPOp — Observability</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
  <script src="https://unpkg.com/@phosphor-icons/web@2.1.1"></script>
  <script>
    tailwind.config = {
      darkMode: 'class',
      theme: {
        extend: {
          colors: {
            neutral: {
              50: '#fafafa',
              100: '#f4f4f5',
              200: '#e4e4e7',
              300: '#d4d4d8',
              400: '#a1a1aa',
              500: '#71717a',
              600: '#52525b',
              700: '#3f3f46',
              800: '#27272a',
              900: '#18181b',
              950: '#09090b',
            }
          }
        }
      }
    }
  </script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    body { font-family: 'Be Vietnam Pro', -apple-system, sans-serif; font-size: 13px; letter-spacing: -0.01em; }
    code, pre, .font-mono { font-family: 'JetBrains Mono', monospace; }
    ::-webkit-scrollbar { width: 4px; height: 4px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: #d4d4d8; border-radius: 9999px; }
    .dark ::-webkit-scrollbar-thumb { background: #3f3f46; }
  </style>
</head>
<body class="bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100 min-h-screen flex flex-col antialiased selection:bg-neutral-900 selection:text-white dark:selection:bg-white dark:selection:text-black">

  <!-- Lean Header Bar -->
  <header class="border-b border-neutral-200 dark:border-neutral-800 bg-white/95 dark:bg-neutral-900/95 backdrop-blur sticky top-0 z-30 px-5 py-2.5 flex items-center justify-between">
    <!-- Brand & Session ID -->
    <div class="flex items-center space-x-3">
      <div class="flex items-center space-x-2">
        <span class="w-2.5 h-2.5 rounded-full bg-neutral-950 dark:bg-white inline-block"></span>
        <span class="font-bold tracking-tight text-neutral-950 dark:text-white text-sm">MCPOp</span>
      </div>
      <span class="text-neutral-300 dark:text-neutral-700 font-mono">/</span>
      <span class="text-[11px] font-mono text-neutral-500 truncate max-w-xs" id="headerCommand">Connecting...</span>
    </div>

    <!-- Actions & Controls -->
    <div class="flex items-center space-x-2.5">
      <!-- Live SSE Status Indicator (Semantic Green) -->
      <div class="flex items-center space-x-1.5 text-xs bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/80 px-2.5 py-1 rounded text-emerald-700 dark:text-emerald-400 font-mono text-[11px] font-medium">
        <span class="relative flex h-2 w-2" aria-hidden="true">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-500 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-600"></span>
        </span>
        <span id="liveStatus">Live</span>
      </div>

      <!-- Session Picker -->
      <div class="flex items-center space-x-1 bg-neutral-100 dark:bg-neutral-800 rounded px-1 py-0.5 border border-neutral-200 dark:border-neutral-700">
        <label for="sessionSelect" class="sr-only">Session</label>
        <select id="sessionSelect" aria-label="MCP Session" class="bg-transparent text-xs text-neutral-800 dark:text-neutral-200 font-medium py-1 px-1.5 focus:outline-none cursor-pointer">
          <option value="">Loading...</option>
        </select>
        <button onclick="refreshData()" aria-label="Refresh" title="Refresh" class="w-6 h-6 flex items-center justify-center text-neutral-500 hover:text-neutral-950 dark:hover:text-white transition">
          <i class="ph ph-arrows-clockwise text-[14px]"></i>
        </button>
      </div>

      <!-- Theme Toggle -->
      <button onclick="toggleTheme()" id="themeBtn" aria-label="Toggle Theme" class="w-7 h-7 rounded border border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-700 flex items-center justify-center transition">
        <i class="ph ph-moon text-[14px]"></i>
      </button>
    </div>
  </header>

  <!-- Main Container -->
  <main class="flex-1 max-w-7xl w-full mx-auto p-5 space-y-4">

    <!-- KPI Strip (Compact, High-Information Density with Semantic Accents) -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 shadow-2xs">
        <div class="flex items-center justify-between text-[11px] text-neutral-500 font-medium uppercase tracking-wider">
          <span>Total Calls</span>
          <i class="ph ph-stack text-[15px] text-neutral-400"></i>
        </div>
        <div class="text-xl font-bold text-neutral-950 dark:text-white mt-1 font-mono" id="statTotalCalls">0</div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 shadow-2xs">
        <div class="flex items-center justify-between text-[11px] text-neutral-500 font-medium uppercase tracking-wider">
          <span>Reliability</span>
          <i class="ph ph-check-circle text-[15px] text-emerald-500"></i>
        </div>
        <div class="text-xl font-bold text-emerald-600 dark:text-emerald-400 mt-1 font-mono" id="statSuccessRate">100%</div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 shadow-2xs">
        <div class="flex items-center justify-between text-[11px] text-neutral-500 font-medium uppercase tracking-wider">
          <span>Avg Latency</span>
          <i class="ph ph-timer text-[15px] text-neutral-400"></i>
        </div>
        <div class="text-xl font-bold text-neutral-950 dark:text-white mt-1 font-mono" id="statAvgLatency">0 ms</div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 shadow-2xs">
        <div class="flex items-center justify-between text-[11px] text-neutral-500 font-medium uppercase tracking-wider">
          <span>Silent Failures</span>
          <i class="ph ph-warning-circle text-[15px] text-rose-500" id="statFailuresIcon"></i>
        </div>
        <div class="text-xl font-bold text-neutral-950 dark:text-white mt-1 font-mono" id="statFailures">0</div>
      </div>
    </div>

    <!-- Design:OS Lean Signal Track Diagram (Inline Topology with Semantic Signals) -->
    <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg px-4 py-2.5 flex flex-wrap items-center justify-between gap-3 text-xs shadow-2xs">
      <div class="flex items-center space-x-2 font-mono">
        <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
        <span class="font-bold text-neutral-950 dark:text-white text-[11px] uppercase tracking-wider">Topology</span>
      </div>

      <div class="flex items-center space-x-2 text-[11px] font-mono text-neutral-600 dark:text-neutral-300 overflow-x-auto py-0.5">
        <div class="flex items-center space-x-1.5 bg-neutral-100 dark:bg-neutral-800 px-2.5 py-1 rounded border border-neutral-200 dark:border-neutral-700">
          <i class="ph ph-robot text-[14px]"></i>
          <span>Client (stdio)</span>
        </div>
        <span class="text-neutral-400 dark:text-neutral-600 font-mono">──►</span>
        <div class="flex items-center space-x-1.5 bg-neutral-950 text-white dark:bg-white dark:text-neutral-950 px-2.5 py-1 rounded font-bold shadow-xs">
          <i class="ph ph-shield-check text-[14px] text-emerald-400 dark:text-emerald-600"></i>
          <span>MCPOp Interceptor</span>
        </div>
        <span class="text-neutral-400 dark:text-neutral-600 font-mono">──►</span>
        <div class="flex items-center space-x-1.5 bg-neutral-100 dark:bg-neutral-800 px-2.5 py-1 rounded border border-neutral-200 dark:border-neutral-700">
          <i class="ph ph-hard-drives text-[14px]"></i>
          <span id="pipelineTarget">Target Server</span>
        </div>
      </div>

      <div class="flex items-center space-x-3 text-[11px] text-neutral-500 font-mono">
        <span>Overhead: <strong class="text-emerald-600 dark:text-emerald-400">&lt;0.5ms</strong></span>
        <span>Spec: <strong class="text-neutral-900 dark:text-neutral-100">JSON-RPC 2.0</strong></span>
      </div>
    </div>

    <!-- Design:OS Visual Telemetry & Analytics Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
      <!-- Latency Timeline Line Chart (2 Cols) -->
      <div class="lg:col-span-2 bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3.5 shadow-2xs">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center space-x-2">
            <i class="ph ph-chart-line-up text-[16px] text-neutral-950 dark:text-white"></i>
            <span class="font-bold text-xs uppercase tracking-wider text-neutral-950 dark:text-white">Execution Latency (ms)</span>
          </div>
          <div class="flex items-center space-x-3 text-[11px] font-mono text-neutral-500">
            <span>p50: <strong id="valP50" class="text-neutral-950 dark:text-white">0ms</strong></span>
            <span>p90: <strong id="valP90" class="text-neutral-950 dark:text-white">0ms</strong></span>
            <span>p99: <strong id="valP99" class="text-neutral-950 dark:text-white">0ms</strong></span>
          </div>
        </div>
        <div class="h-32 w-full">
          <canvas id="latencyChart" role="img" aria-label="Latency Timeline"></canvas>
        </div>
      </div>

      <!-- Tool Share Doughnut Chart (1 Col) -->
      <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3.5 shadow-2xs flex flex-col justify-between">
        <div class="flex items-center justify-between mb-1">
          <div class="flex items-center space-x-2">
            <i class="ph ph-chart-pie-slice text-[16px] text-neutral-950 dark:text-white"></i>
            <span class="font-bold text-xs uppercase tracking-wider text-neutral-950 dark:text-white">Tool Invocations</span>
          </div>
          <span class="text-[11px] font-mono text-neutral-500" id="distinctToolCount">0 tools</span>
        </div>
        <div class="h-32 w-full flex items-center justify-center">
          <canvas id="toolPieChart" role="img" aria-label="Tool Invocation Share"></canvas>
        </div>
      </div>
    </div>

    <!-- Failure Alerts Strip (Semantic Colors: Red for Loops, Amber for Schema Mismatches, Orange for Slow Tools) -->
    <div id="failureAlertsContainer" class="hidden space-y-2" role="region" aria-label="Failure Alerts">
      <div id="failureAlertsList" class="space-y-2"></div>
    </div>

    <!-- Realtime Traces Waterfall Table -->
    <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden shadow-2xs">
      
      <!-- Table Control Toolbar -->
      <div class="p-3 border-b border-neutral-200 dark:border-neutral-800 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center space-x-2">
          <span class="font-bold text-xs uppercase tracking-wider text-neutral-950 dark:text-white">Live Tool Traces</span>
          <span id="traceCountBadge" class="text-[11px] font-mono text-neutral-500 bg-neutral-100 dark:bg-neutral-800 px-2 py-0.5 rounded font-semibold">0</span>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Filter Tabs -->
          <div class="flex items-center bg-neutral-100 dark:bg-neutral-800 p-0.5 rounded text-xs border border-neutral-200 dark:border-neutral-700" role="group">
            <button onclick="setFilter('all')" id="filterAll" class="px-2.5 py-1 rounded bg-white dark:bg-neutral-700 shadow-2xs text-neutral-950 dark:text-white font-bold">All</button>
            <button onclick="setFilter('errors')" id="filterErrors" class="px-2.5 py-1 rounded text-neutral-600 dark:text-neutral-400 hover:text-rose-600 dark:hover:text-rose-400 font-medium">Errors</button>
            <button onclick="setFilter('slow')" id="filterSlow" class="px-2.5 py-1 rounded text-neutral-600 dark:text-neutral-400 hover:text-amber-600 dark:hover:text-amber-400 font-medium">Slow</button>
          </div>

          <!-- Search -->
          <div class="relative">
            <input id="searchTraces" oninput="renderTraces()" type="text" placeholder="Search tools..." aria-label="Search tools" class="bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-700 text-xs rounded px-2.5 py-1 text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white w-40">
          </div>
        </div>
      </div>

      <!-- Traces Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs" role="table">
          <thead class="bg-neutral-50 dark:bg-neutral-950/80 text-neutral-500 dark:text-neutral-400 border-b border-neutral-200 dark:border-neutral-800 font-semibold">
            <tr>
              <th scope="col" class="py-2.5 px-4 w-24">Status</th>
              <th scope="col" class="py-2.5 px-4 w-44">Tool Name</th>
              <th scope="col" class="py-2.5 px-4">Arguments Payload</th>
              <th scope="col" class="py-2.5 px-4 w-32">Latency</th>
              <th scope="col" class="py-2.5 px-4 w-28">Timestamp</th>
              <th scope="col" class="py-2.5 px-4 text-right w-24">Actions</th>
            </tr>
          </thead>
          <tbody id="traceTableBody" class="divide-y divide-neutral-100 dark:divide-neutral-800/80 font-mono">
            <tr>
              <td colspan="6" class="py-10 text-center text-neutral-400 font-sans">No tool calls recorded yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </main>

  <!-- Tool Detail & Replay Drawer Modal -->
  <div id="detailModal" onclick="handleBackdropClick(event)" class="fixed inset-0 bg-neutral-950/50 backdrop-blur-2xs z-50 hidden flex items-center justify-center p-4" role="dialog" aria-modal="true" aria-labelledby="modalToolName">
    <div id="modalCard" class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-xl max-w-2xl w-full shadow-2xl flex flex-col max-h-[85vh] overflow-hidden">
      
      <!-- Modal Header -->
      <div class="px-5 py-3.5 border-b border-neutral-200 dark:border-neutral-800 flex items-center justify-between">
        <div class="flex items-center space-x-2">
          <h3 class="font-bold text-neutral-950 dark:text-white text-sm font-mono" id="modalToolName">tool/name</h3>
          <span id="modalStatusBadge" class="text-[11px] font-bold px-2 py-0.5 rounded uppercase font-mono"></span>
        </div>
        <button onclick="closeModal()" aria-label="Close" class="w-6 h-6 rounded flex items-center justify-center text-neutral-400 hover:text-neutral-950 dark:hover:text-white">
          <i class="ph ph-x text-[16px]"></i>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-5 overflow-y-auto space-y-4 text-xs">
        <div>
          <div class="flex items-center justify-between mb-1">
            <label for="modalArgsInput" class="block text-neutral-500 font-bold uppercase tracking-wider text-[11px]">Arguments (JSON)</label>
            <button onclick="copyArgsJSON()" class="text-[11px] text-neutral-950 dark:text-white hover:underline flex items-center space-x-1 font-sans font-semibold">
              <i class="ph ph-copy text-[14px]"></i>
              <span id="copyArgsText">Copy</span>
            </button>
          </div>
          <textarea id="modalArgsInput" aria-label="Arguments JSON" class="w-full bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-800 rounded-lg p-2.5 font-mono text-neutral-900 dark:text-neutral-100 text-xs focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white h-28"></textarea>
        </div>

        <div>
          <label class="block text-neutral-500 font-bold mb-1 uppercase tracking-wider text-[11px]">Response Result</label>
          <pre id="modalResultText" class="w-full bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-800 rounded-lg p-2.5 font-mono text-neutral-800 dark:text-neutral-200 text-xs overflow-x-auto max-h-40 whitespace-pre-wrap"></pre>
        </div>

        <!-- Replay Result Box -->
        <div id="replayResultBox" class="hidden p-3 rounded-lg border border-neutral-300 dark:border-neutral-700 bg-neutral-100 dark:bg-neutral-800 space-y-1.5">
          <div class="flex items-center justify-between">
            <span class="font-bold text-neutral-900 dark:text-white text-xs">Replay Execution</span>
            <span id="replayLatencyBadge" class="text-[11px] font-mono text-neutral-600 dark:text-neutral-300 font-semibold"></span>
          </div>
          <pre id="replayOutputText" class="font-mono text-xs text-neutral-900 dark:text-neutral-100 whitespace-pre-wrap bg-white dark:bg-neutral-950 p-2.5 rounded border border-neutral-200 dark:border-neutral-700"></pre>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-5 py-3 border-t border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-950 flex items-center justify-between">
        <div class="text-[11px] text-neutral-500 font-mono" id="modalMetadata">Latency: 0ms</div>
        <div class="flex items-center space-x-2">
          <button onclick="closeModal()" class="px-3.5 py-1.5 rounded border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 hover:bg-neutral-100 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 text-xs font-semibold">Close</button>
          <button id="btnReplay" onclick="runReplay()" class="px-3.5 py-1.5 rounded bg-neutral-950 hover:bg-neutral-800 dark:bg-white dark:hover:bg-neutral-200 text-white dark:text-neutral-950 text-xs font-bold flex items-center space-x-1.5">
            <i class="ph ph-play text-[14px]"></i>
            <span>1-Click Replay</span>
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- JavaScript Logic -->
  <script>
    let currentSessionId = '';
    let sessions = [];
    let traces = [];
    let failures = [];
    let currentModalTrace = null;
    let currentFilter = 'all';
    let latencyChartInstance = null;
    let toolPieChartInstance = null;

    function esc(value) {
      return String(value == null ? '' : value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
    }

    function toggleTheme() {
      const html = document.documentElement;
      if (html.classList.contains('dark')) {
        html.classList.remove('dark');
        html.classList.add('light');
        document.getElementById('themeBtn').innerHTML = '<i class="ph ph-moon text-[14px]"></i>';
        localStorage.setItem('theme', 'light');
      } else {
        html.classList.remove('light');
        html.classList.add('dark');
        document.getElementById('themeBtn').innerHTML = '<i class="ph ph-sun text-[14px]"></i>';
        localStorage.setItem('theme', 'dark');
      }
      renderCharts();
    }

    if (localStorage.getItem('theme') === 'dark') {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    }

    async function init() {
      await fetchSessions();
      setupSSE();

      window.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') closeModal();
      });
    }

    async function fetchSessions() {
      try {
        const res = await fetch('/api/sessions');
        sessions = await res.json();
        const select = document.getElementById('sessionSelect');
        select.innerHTML = '';

        if (!sessions || sessions.length === 0) {
          select.innerHTML = '<option value="">No sessions</option>';
          return;
        }

        sessions.forEach((s) => {
          const opt = document.createElement('option');
          opt.value = s.id;
          opt.textContent = s.command;
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
        renderCharts();
      } catch (err) {
        console.error('Failed to refresh data:', err);
      }
    }

    function renderStats(stats) {
      if (!stats) return;
      document.getElementById('statTotalCalls').textContent = stats.total_calls || 0;
      
      const succRate = stats.success_rate || 100;
      const rateEl = document.getElementById('statSuccessRate');
      rateEl.textContent = succRate.toFixed(0) + '%';
      if (succRate < 80) {
        rateEl.className = 'text-xl font-bold text-rose-600 dark:text-rose-400 mt-1 font-mono';
      } else if (succRate < 95) {
        rateEl.className = 'text-xl font-bold text-amber-600 dark:text-amber-400 mt-1 font-mono';
      } else {
        rateEl.className = 'text-xl font-bold text-emerald-600 dark:text-emerald-400 mt-1 font-mono';
      }

      document.getElementById('statAvgLatency').textContent = (stats.avg_latency_ms || 0) + ' ms';
      
      const failCount = stats.failure_count || 0;
      const failEl = document.getElementById('statFailures');
      failEl.textContent = failCount;
      if (failCount > 0) {
        failEl.className = 'text-xl font-bold text-rose-600 dark:text-rose-400 mt-1 font-mono';
      } else {
        failEl.className = 'text-xl font-bold text-neutral-950 dark:text-white mt-1 font-mono';
      }

      const curr = sessions.find(s => s.id === currentSessionId);
      if (curr) {
        document.getElementById('headerCommand').textContent = curr.command;
        document.getElementById('pipelineTarget').textContent = curr.command.split(' ')[0] || 'Target Server';
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
        let icon = 'ph-repeat';
        let alertColor = 'bg-rose-50 dark:bg-rose-950/40 text-rose-800 dark:text-rose-300 border-rose-200 dark:border-rose-800/80';
        if (f.failure_type === 'schema_mismatch') {
          icon = 'ph-code';
          alertColor = 'bg-amber-50 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 border-amber-200 dark:border-amber-800/80';
        } else if (f.failure_type === 'slow_tool') {
          icon = 'ph-hourglass';
          alertColor = 'bg-orange-50 dark:bg-orange-950/40 text-orange-800 dark:text-orange-300 border-orange-200 dark:border-orange-800/80';
        }

        div.className = 'p-3 rounded-lg border flex items-start space-x-3 shadow-2xs ' + alertColor;
        const iconEl = document.createElement('i');
        iconEl.className = 'ph-bold ' + icon + ' mt-0.5 text-[16px]';

        const wrap = document.createElement('div');
        wrap.className = 'flex-1';

        const head = document.createElement('div');
        head.className = 'flex items-center justify-between';

        const typeEl = document.createElement('span');
        typeEl.className = 'font-bold text-[11px] uppercase tracking-wider font-mono';
        typeEl.textContent = String(f.failure_type || '').replace('_', ' ');

        const timeEl = document.createElement('span');
        timeEl.className = 'text-[11px] font-mono opacity-70';
        timeEl.textContent = new Date(f.created_at).toLocaleTimeString();

        const desc = document.createElement('p');
        desc.className = 'text-xs mt-0.5 font-medium';
        desc.textContent = f.description || '';

        head.appendChild(typeEl);
        head.appendChild(timeEl);
        wrap.appendChild(head);
        wrap.appendChild(desc);
        div.appendChild(iconEl);
        div.appendChild(wrap);
        list.appendChild(div);
      });
    }

    function setFilter(f) {
      currentFilter = f;
      ['All', 'Errors', 'Slow'].forEach(name => {
        const btn = document.getElementById('filter' + name);
        if (name.toLowerCase() === f) {
          btn.className = 'px-2.5 py-1 rounded bg-white dark:bg-neutral-700 shadow-2xs text-neutral-950 dark:text-white font-bold';
        } else {
          btn.className = 'px-2.5 py-1 rounded text-neutral-600 dark:text-neutral-400 hover:text-neutral-950 dark:hover:text-white font-medium';
        }
      });
      renderTraces();
    }

    function renderTraces() {
      const tbody = document.getElementById('traceTableBody');
      const search = document.getElementById('searchTraces').value.toLowerCase();
      document.getElementById('traceCountBadge').textContent = traces.length;

      let filtered = traces.filter(t => t.tool_name.toLowerCase().includes(search));
      if (currentFilter === 'errors') {
        filtered = filtered.filter(t => t.is_error || t.status === 'failed');
      } else if (currentFilter === 'slow') {
        filtered = filtered.filter(t => t.latency_ms >= 1000);
      }

      if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="py-10 text-center text-neutral-400 font-sans">No matching tool calls.</td></tr>';
        return;
      }

      tbody.innerHTML = '';
      filtered.forEach(t => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-neutral-100/60 dark:hover:bg-neutral-800/40 transition cursor-pointer';

        // Semantic Status Badges
        let statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">SUCCESS</span>';
        if (t.is_error || t.status === 'failed') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-rose-100 dark:bg-rose-950/60 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-800">FAILED</span>';
        } else if (t.status === 'pending') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-indigo-100 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800 animate-pulse">RUNNING</span>';
        }

        let argsPreview = t.arguments;
        if (argsPreview && argsPreview.length > 50) {
          argsPreview = argsPreview.substring(0, 50) + '...';
        }

        // Semantic Latency Bar & Colors
        let latColor = 'text-emerald-600 dark:text-emerald-400';
        let barColor = 'bg-emerald-500';
        let barWidth = Math.min(100, Math.max(8, (t.latency_ms / 3000) * 100));

        if (t.latency_ms > 1000) {
          latColor = 'text-amber-600 dark:text-amber-400';
          barColor = 'bg-amber-500';
        }
        if (t.latency_ms >= 3000) {
          latColor = 'text-rose-600 dark:text-rose-400 font-bold';
          barColor = 'bg-rose-500';
        }

        tr.innerHTML =
          '<td class="py-2.5 px-4">' + statusBadge + '</td>' +
          '<td class="py-2.5 px-4 font-bold text-neutral-950 dark:text-white">' + esc(t.tool_name) + '</td>' +
          '<td class="py-2.5 px-4 text-neutral-500 truncate max-w-sm">' + esc(argsPreview) + '</td>' +
          '<td class="py-2.5 px-4 font-mono">' +
            '<div class="flex items-center space-x-2">' +
              '<span class="font-bold ' + latColor + '">' + esc(t.latency_ms) + 'ms</span>' +
              '<div class="w-12 bg-neutral-200 dark:bg-neutral-800 h-1.5 rounded-full overflow-hidden" aria-hidden="true">' +
                '<div class="' + barColor + ' h-full rounded-full" style="width: ' + barWidth + '%"></div>' +
              '</div>' +
            '</div>' +
          '</td>' +
          '<td class="py-2.5 px-4 text-neutral-400 text-[11px]">' + esc(new Date(t.created_at).toLocaleTimeString()) + '</td>' +
          '<td class="py-2.5 px-4 text-right space-x-1 font-sans"></td>';

        const inspectBtn = document.createElement('button');
        inspectBtn.type = 'button';
        inspectBtn.setAttribute('aria-label', 'Inspect');
        inspectBtn.className = 'px-2.5 py-1 rounded border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 hover:bg-neutral-100 dark:hover:bg-neutral-700 text-neutral-900 dark:text-neutral-100 text-[11px] font-semibold';
        inspectBtn.textContent = 'Inspect';
        inspectBtn.addEventListener('click', (e) => {
          e.stopPropagation();
          inspectTrace(t.id);
        });
        tr.lastElementChild.appendChild(inspectBtn);

        tr.onclick = (e) => {
          if (e.target.tagName !== 'BUTTON') inspectTrace(t.id);
        };
        tbody.appendChild(tr);
      });
    }

    function renderCharts() {
      const isDark = document.documentElement.classList.contains('dark');
      const gridColor = isDark ? 'rgba(63, 63, 70, 0.3)' : 'rgba(228, 228, 231, 0.8)';
      const textColor = isDark ? '#a1a1aa' : '#71717a';
      const lineColor = isDark ? '#ffffff' : '#18181b';

      // Percentiles
      const latencies = traces.map(t => t.latency_ms).sort((a, b) => a - b);
      if (latencies.length > 0) {
        const p50 = latencies[Math.floor(latencies.length * 0.50)] || 0;
        const p90 = latencies[Math.floor(latencies.length * 0.90)] || 0;
        const p99 = latencies[Math.floor(latencies.length * 0.99)] || 0;
        document.getElementById('valP50').textContent = p50 + 'ms';
        document.getElementById('valP90').textContent = p90 + 'ms';
        document.getElementById('valP99').textContent = p99 + 'ms';
      }

      // 1. Latency Timeline
      const orderedTraces = [...traces].reverse();
      const labels = orderedTraces.map((t, i) => '#' + (i + 1));
      const latValues = orderedTraces.map(t => t.latency_ms);

      const ctx1 = document.getElementById('latencyChart').getContext('2d');
      if (latencyChartInstance) latencyChartInstance.destroy();

      latencyChartInstance = new Chart(ctx1, {
        type: 'line',
        data: {
          labels: labels,
          datasets: [{
            data: latValues,
            borderColor: lineColor,
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(24, 24, 27, 0.04)',
            borderWidth: 1.5,
            fill: true,
            tension: 0.25,
            pointRadius: 2.5,
            pointBackgroundColor: lineColor
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: { legend: { display: false } },
          scales: {
            x: {
              grid: { color: gridColor },
              ticks: { color: textColor, font: { family: 'Be Vietnam Pro', size: 10 } }
            },
            y: {
              grid: { color: gridColor },
              ticks: { color: textColor, font: { family: 'Be Vietnam Pro', size: 10 } },
              beginAtZero: true
            }
          }
        }
      });

      // 2. Tool Share Doughnut
      const toolCounts = {};
      traces.forEach(t => {
        toolCounts[t.tool_name] = (toolCounts[t.tool_name] || 0) + 1;
      });

      const pieLabels = Object.keys(toolCounts);
      const pieData = Object.values(toolCounts);
      document.getElementById('distinctToolCount').textContent = pieLabels.length + ' distinct';

      const ctx2 = document.getElementById('toolPieChart').getContext('2d');
      if (toolPieChartInstance) toolPieChartInstance.destroy();

      toolPieChartInstance = new Chart(ctx2, {
        type: 'doughnut',
        data: {
          labels: pieLabels,
          datasets: [{
            data: pieData,
            backgroundColor: isDark 
              ? ['#ffffff', '#d4d4d8', '#a1a1aa', '#71717a', '#3f3f46']
              : ['#18181b', '#3f3f46', '#71717a', '#a1a1aa', '#e4e4e7'],
            borderWidth: 1.5,
            borderColor: isDark ? '#09090b' : '#ffffff'
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'right',
              labels: { color: textColor, boxWidth: 8, font: { family: 'Be Vietnam Pro', size: 10 } }
            }
          },
          cutout: '72%'
        }
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
        statusBadge.className = 'text-[11px] font-bold px-2.5 py-0.5 rounded uppercase bg-rose-100 text-rose-700 border border-rose-200 font-mono';
      } else {
        statusBadge.textContent = 'Success';
        statusBadge.className = 'text-[11px] font-bold px-2.5 py-0.5 rounded uppercase bg-emerald-100 text-emerald-700 border border-emerald-200 font-mono';
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

    function handleBackdropClick(e) {
      if (e.target.id === 'detailModal') closeModal();
    }

    function copyArgsJSON() {
      const text = document.getElementById('modalArgsInput').value;
      navigator.clipboard.writeText(text).then(() => {
        const label = document.getElementById('copyArgsText');
        label.textContent = 'Copied!';
        setTimeout(() => { label.textContent = 'Copy'; }, 2000);
      });
    }

    async function runReplay() {
      if (!currentModalTrace) return;
      const btn = document.getElementById('btnReplay');
      const box = document.getElementById('replayResultBox');
      const out = document.getElementById('replayOutputText');
      const lat = document.getElementById('replayLatencyBadge');

      btn.disabled = true;
      btn.innerHTML = '<i class="ph ph-spinner animate-spin text-[14px]"></i> <span>Replaying...</span>';

      let parsedArgs = {};
      try {
        parsedArgs = JSON.parse(document.getElementById('modalArgsInput').value);
      } catch (e) {
        alert('Invalid JSON in arguments');
        btn.disabled = false;
        btn.innerHTML = '<i class="ph ph-play text-[14px]"></i> <span>1-Click Replay</span>';
        return;
      }

      try {
        const res = await fetch('/api/replay', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            session_id: currentSessionId,
            tool_name: currentModalTrace.tool_name,
            arguments: parsedArgs
          })
        });

        const data = await res.json();
        box.classList.remove('hidden');
        lat.textContent = data.latency_ms + 'ms';
        out.textContent = JSON.stringify(data.result || data.error, null, 2);
      } catch (err) {
        box.classList.remove('hidden');
        lat.textContent = 'Error';
        out.textContent = 'Replay failed: ' + err.message;
      } finally {
        btn.disabled = false;
        btn.innerHTML = '<i class="ph ph-play text-[14px]"></i> <span>1-Click Replay</span>';
      }
    }

    function setupSSE() {
      const evtSource = new EventSource('/api/events');
      evtSource.onopen = () => {
        document.getElementById('liveStatus').textContent = 'Live';
      };
      evtSource.onerror = () => {
        document.getElementById('liveStatus').textContent = 'Reconnecting...';
      };
      evtSource.onmessage = () => { try { refreshData(); } catch {} };
      evtSource.addEventListener('tool_call', () => { refreshData(); });
      evtSource.addEventListener('failure', () => { refreshData(); });
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
