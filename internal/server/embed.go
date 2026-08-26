package server

import (
	_ "embed"
	"net/http"
)

// Embedded modern Light Theme Dashboard with Design:OS Visuals, Charts & Diagrams
const DashboardHTML = `<!DOCTYPE html>
<html lang="en" class="light">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MCPOp — MCP Observability & Failure Catcher</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
  <script>
    tailwind.config = {
      darkMode: 'class',
      theme: {
        extend: {
          colors: {
            brand: {
              50: '#f5f3ff',
              100: '#ede9fe',
              500: '#6366f1',
              600: '#4f46e5',
              700: '#4338ca',
            }
          }
        }
      }
    }
  </script>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  <style>
    body { font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif; }
    code, pre, .font-mono { font-family: 'JetBrains Mono', monospace; }
    ::-webkit-scrollbar { width: 6px; height: 6px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 9999px; }
    .dark ::-webkit-scrollbar-thumb { background: #334155; }
    
    /* Background subtle grid pattern */
    .bg-grid-pattern {
      background-size: 32px 32px;
      background-image: 
        linear-gradient(to right, rgba(226, 232, 240, 0.6) 1px, transparent 1px),
        linear-gradient(to bottom, rgba(226, 232, 240, 0.6) 1px, transparent 1px);
    }
    .dark .bg-grid-pattern {
      background-image: 
        linear-gradient(to right, rgba(30, 41, 59, 0.5) 1px, transparent 1px),
        linear-gradient(to bottom, rgba(30, 41, 59, 0.5) 1px, transparent 1px);
    }
  </style>
</head>
<body class="bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 min-h-screen flex flex-col antialiased transition-colors duration-200 bg-grid-pattern selection:bg-indigo-500 selection:text-white">

  <!-- Top Navigation Bar -->
  <header class="border-b border-slate-200/80 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 backdrop-blur sticky top-0 z-30 px-6 py-3.5 flex items-center justify-between shadow-xs">
    <div class="flex items-center space-x-3.5">
      <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center shadow-md shadow-indigo-500/20 text-white font-bold">
        <i class="fa-solid fa-bolt text-sm"></i>
      </div>
      <div>
        <div class="flex items-center space-x-2">
          <span class="font-extrabold tracking-tight text-slate-900 dark:text-white text-base">MCPOp</span>
          <span class="text-[11px] font-semibold bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800/60 px-2 py-0.5 rounded-full">Observability</span>
          <span class="text-[10px] font-medium text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded">v0.1.0</span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400">Silent Failure Catcher & Realtime Waterfall for MCP</p>
      </div>
    </div>

    <!-- Active Session, SSE Pulse & Theme Toggle -->
    <div class="flex items-center space-x-3">
      <!-- Live SSE Connection Pill -->
      <div class="flex items-center space-x-2 text-xs bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/60 px-3 py-1.5 rounded-lg text-emerald-700 dark:text-emerald-400 shadow-2xs">
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-500 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-600"></span>
        </span>
        <span id="liveStatus" class="font-medium">Live SSE Connected</span>
      </div>

      <!-- Session Picker -->
      <div class="flex items-center space-x-1.5 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-1 shadow-2xs">
        <i class="fa-solid fa-terminal text-xs text-slate-400 pl-2"></i>
        <select id="sessionSelect" class="bg-transparent text-xs text-slate-700 dark:text-slate-200 font-medium py-1 px-2 focus:outline-none cursor-pointer">
          <option value="">Loading sessions...</option>
        </select>
        <button onclick="refreshData()" title="Refresh" class="p-1 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 transition">
          <i class="fa-solid fa-arrows-rotate text-xs"></i>
        </button>
      </div>

      <!-- Theme Switcher -->
      <button onclick="toggleTheme()" id="themeBtn" class="w-8 h-8 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 flex items-center justify-center transition shadow-2xs">
        <i class="fa-solid fa-moon text-xs"></i>
      </button>
    </div>
  </header>

  <!-- Main Container -->
  <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">

    <!-- Interactive Architecture Flow Diagram (Design:OS Flow) -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200/90 dark:border-slate-800 rounded-2xl p-5 shadow-xs">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center space-x-2 text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
          <i class="fa-solid fa-network-wired text-indigo-600"></i>
          <span>Transparent Interceptor Topology</span>
        </div>
        <span class="text-[11px] text-slate-500 dark:text-slate-400 font-mono" id="diagCommand">Command: (waiting for session)</span>
      </div>

      <!-- Visual Flow Chart Diagram -->
      <div class="grid grid-cols-1 md:grid-cols-5 gap-3 items-center text-center">
        <!-- Node 1: AI Client -->
        <div class="bg-slate-50 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-xl p-3.5 shadow-2xs">
          <div class="w-8 h-8 mx-auto rounded-lg bg-indigo-100 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400 flex items-center justify-center mb-1.5">
            <i class="fa-solid fa-robot text-sm"></i>
          </div>
          <div class="text-xs font-bold text-slate-800 dark:text-slate-200">AI Client</div>
          <div class="text-[10px] text-slate-500">Claude / Cursor / Agent</div>
        </div>

        <!-- Connection Arrow 1 -->
        <div class="flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 py-1">
          <span class="text-[10px] font-mono font-medium text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/80 px-2 py-0.5 rounded-full border border-indigo-200/60 dark:border-indigo-800 mb-1">stdio pipe</span>
          <i class="fa-solid fa-arrow-right-arrow-left text-xs hidden md:block"></i>
          <i class="fa-solid fa-arrow-down-up text-xs md:hidden"></i>
        </div>

        <!-- Node 2: MCPOp Engine (Hero Focus) -->
        <div class="bg-gradient-to-b from-indigo-50/80 to-white dark:from-slate-800 dark:to-slate-900 border-2 border-indigo-500/80 rounded-xl p-3.5 shadow-sm shadow-indigo-500/10 relative">
          <div class="absolute -top-2.5 left-1/2 -translate-x-1/2 bg-indigo-600 text-white text-[9px] font-extrabold uppercase px-2 py-0.5 rounded-full tracking-wider">
            Active Proxy
          </div>
          <div class="w-8 h-8 mx-auto rounded-lg bg-indigo-600 text-white flex items-center justify-center mb-1.5 shadow-xs">
            <i class="fa-solid fa-shield-halved text-sm"></i>
          </div>
          <div class="text-xs font-bold text-indigo-950 dark:text-white">MCPOp Core</div>
          <div class="text-[10px] text-indigo-600 dark:text-indigo-400 font-mono">&lt;1ms Overhead Interceptor</div>
        </div>

        <!-- Connection Arrow 2 -->
        <div class="flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 py-1">
          <span class="text-[10px] font-mono font-medium text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/80 px-2 py-0.5 rounded-full border border-emerald-200/60 dark:border-emerald-800 mb-1">JSON-RPC 2.0</span>
          <i class="fa-solid fa-arrow-right-arrow-left text-xs hidden md:block"></i>
          <i class="fa-solid fa-arrow-down-up text-xs md:hidden"></i>
        </div>

        <!-- Node 3: Target Server -->
        <div class="bg-slate-50 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-xl p-3.5 shadow-2xs">
          <div class="w-8 h-8 mx-auto rounded-lg bg-emerald-100 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mb-1.5">
            <i class="fa-solid fa-server text-sm"></i>
          </div>
          <div class="text-xs font-bold text-slate-800 dark:text-slate-200">Target MCP Server</div>
          <div class="text-[10px] text-slate-500">Python / Node / Go Tools</div>
        </div>
      </div>
    </div>

    <!-- Stats KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-4 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs font-semibold">
          <span>Total Tool Invocations</span>
          <div class="w-7 h-7 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 flex items-center justify-center">
            <i class="fa-solid fa-layer-group text-xs"></i>
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900 dark:text-white mt-2" id="statTotalCalls">0</div>
        <div class="text-[11px] text-slate-500 mt-1 flex items-center space-x-1">
          <i class="fa-solid fa-chart-simple text-[10px] text-indigo-500"></i>
          <span>Captured in real-time</span>
        </div>
      </div>

      <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-4 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs font-semibold">
          <span>Success Reliability</span>
          <div class="w-7 h-7 rounded-lg bg-emerald-50 dark:bg-emerald-950/60 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
            <i class="fa-solid fa-circle-check text-xs"></i>
          </div>
        </div>
        <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-2" id="statSuccessRate">100%</div>
        <div class="text-[11px] text-slate-500 mt-1 flex items-center space-x-1" id="statErrorCallsBox">
          <span id="statErrorCalls">0 errors</span>
        </div>
      </div>

      <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-4 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs font-semibold">
          <span>Average Latency</span>
          <div class="w-7 h-7 rounded-lg bg-amber-50 dark:bg-amber-950/60 text-amber-600 dark:text-amber-400 flex items-center justify-center">
            <i class="fa-solid fa-stopwatch text-xs"></i>
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900 dark:text-white mt-2" id="statAvgLatency">0 ms</div>
        <div class="text-[11px] text-slate-500 mt-1 flex items-center space-x-1">
          <span>Execution time per call</span>
        </div>
      </div>

      <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-4 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition">
        <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs font-semibold">
          <span>Silent Failures Intercepted</span>
          <div class="w-7 h-7 rounded-lg bg-rose-50 dark:bg-rose-950/60 text-rose-600 dark:text-rose-400 flex items-center justify-center">
            <i class="fa-solid fa-triangle-exclamation text-xs"></i>
          </div>
        </div>
        <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-2" id="statFailures">0</div>
        <div class="text-[11px] text-slate-500 mt-1 flex items-center space-x-1">
          <span>Loops, Schemas & Slow Tools</span>
        </div>
      </div>
    </div>

    <!-- Charts & Anomaly Section (Design:OS Visual Analytics) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Chart 1: Latency Waterfall Trend (2 cols) -->
      <div class="lg:col-span-2 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-5 shadow-xs">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center space-x-2">
              <i class="fa-solid fa-chart-line text-indigo-600"></i>
              <span>Tool Call Latency Timeline (ms)</span>
            </h3>
            <p class="text-[11px] text-slate-500">Execution latency distribution with threshold detection</p>
          </div>
          <span class="text-[10px] font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-0.5 rounded">Realtime</span>
        </div>
        <div class="relative h-48 w-full">
          <canvas id="latencyChart"></canvas>
        </div>
      </div>

      <!-- Chart 2: Tool Distribution & Status (1 col) -->
      <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl p-5 shadow-xs flex flex-col justify-between">
        <div>
          <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center space-x-2 mb-1">
            <i class="fa-solid fa-chart-pie text-violet-600"></i>
            <span>Tool Invocation Shares</span>
          </h3>
          <p class="text-[11px] text-slate-500 mb-3">Volume distribution across tool types</p>
        </div>
        <div class="relative h-44 flex items-center justify-center">
          <canvas id="toolPieChart"></canvas>
        </div>
      </div>
    </div>

    <!-- Failure Alerts Section -->
    <div id="failureAlertsContainer" class="hidden space-y-2.5">
      <div class="flex items-center space-x-2 text-xs font-bold uppercase tracking-wider text-rose-600 dark:text-rose-400">
        <i class="fa-solid fa-radiation text-rose-500"></i>
        <span>Heuristic Anomaly Detections (Immediate Action Required)</span>
      </div>
      <div id="failureAlertsList" class="space-y-2"></div>
    </div>

    <!-- Traces Waterfall Table -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl overflow-hidden shadow-xs">
      <div class="p-4 border-b border-slate-200/80 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center space-x-3">
          <h2 class="text-sm font-bold text-slate-900 dark:text-white flex items-center space-x-2">
            <span>Live Tool Trace Waterfall</span>
          </h2>
          <span id="traceCountBadge" class="text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-0.5 rounded-full font-mono font-medium">0 traces</span>
        </div>

        <div class="flex items-center space-x-2.5">
          <!-- Filter Buttons -->
          <div class="flex items-center bg-slate-100 dark:bg-slate-800 p-0.5 rounded-lg text-xs font-medium text-slate-600 dark:text-slate-300">
            <button onclick="setFilter('all')" id="filterAll" class="px-2.5 py-1 rounded-md bg-white dark:bg-slate-700 shadow-2xs text-slate-900 dark:text-white font-semibold">All</button>
            <button onclick="setFilter('errors')" id="filterErrors" class="px-2.5 py-1 rounded-md hover:text-slate-900 dark:hover:text-white">Errors</button>
            <button onclick="setFilter('slow')" id="filterSlow" class="px-2.5 py-1 rounded-md hover:text-slate-900 dark:hover:text-white">Slow</button>
          </div>

          <div class="relative">
            <i class="fa-solid fa-magnifying-glass absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs"></i>
            <input id="searchTraces" oninput="renderTraces()" type="text" placeholder="Search tools..." class="bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-700 text-xs rounded-lg pl-8 pr-3 py-1.5 text-slate-800 dark:text-slate-200 placeholder-slate-400 focus:outline-none focus:border-indigo-500 w-44">
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50/70 dark:bg-slate-950/60 text-slate-500 dark:text-slate-400 border-b border-slate-200/80 dark:border-slate-800 font-semibold">
            <tr>
              <th class="py-3 px-4">Status</th>
              <th class="py-3 px-4">Tool Name</th>
              <th class="py-3 px-4">Arguments Payload</th>
              <th class="py-3 px-4">Latency (Waterfall)</th>
              <th class="py-3 px-4">Timestamp</th>
              <th class="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody id="traceTableBody" class="divide-y divide-slate-100 dark:divide-slate-800 font-mono">
            <tr>
              <td colspan="6" class="py-12 text-center text-slate-400 font-sans">No tool calls recorded in this session yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </main>

  <!-- Tool Detail & Replay Drawer Modal -->
  <div id="detailModal" class="fixed inset-0 bg-slate-900/50 backdrop-blur-xs z-50 hidden flex items-center justify-center p-4">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl max-w-3xl w-full shadow-2xl flex flex-col max-h-[90vh] overflow-hidden">
      
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900">
        <div class="flex items-center space-x-2.5">
          <div class="w-3 h-3 rounded-full bg-indigo-600"></div>
          <h3 class="font-bold text-slate-900 dark:text-white text-base font-mono" id="modalToolName">tool/name</h3>
          <span id="modalStatusBadge" class="text-[10px] font-bold px-2 py-0.5 rounded-full uppercase"></span>
        </div>
        <button onclick="closeModal()" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition">
          <i class="fa-solid fa-xmark text-lg"></i>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 overflow-y-auto space-y-5 text-xs">
        <div>
          <label class="block text-slate-600 dark:text-slate-400 font-bold mb-1.5 uppercase tracking-wider text-[10px]">Input Arguments (Editable JSON)</label>
          <textarea id="modalArgsInput" class="w-full bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl p-3 font-mono text-slate-800 dark:text-slate-200 text-xs focus:outline-none focus:border-indigo-500 h-32"></textarea>
        </div>

        <div>
          <label class="block text-slate-600 dark:text-slate-400 font-bold mb-1.5 uppercase tracking-wider text-[10px]">Execution Result Payload</label>
          <pre id="modalResultText" class="w-full bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl p-3 font-mono text-emerald-600 dark:text-emerald-400 text-xs overflow-x-auto max-h-48 whitespace-pre-wrap"></pre>
        </div>

        <!-- Replay Result Box -->
        <div id="replayResultBox" class="hidden p-4 rounded-xl border border-indigo-200 dark:border-indigo-800/80 bg-indigo-50/60 dark:bg-indigo-950/30 space-y-2">
          <div class="flex items-center justify-between">
            <span class="font-bold text-indigo-900 dark:text-indigo-300 text-xs">Replay Execution Response</span>
            <span id="replayLatencyBadge" class="text-[11px] font-mono font-bold text-indigo-700 dark:text-indigo-400"></span>
          </div>
          <pre id="replayOutputText" class="font-mono text-xs text-slate-800 dark:text-slate-200 whitespace-pre-wrap bg-white dark:bg-slate-950 p-3 rounded-lg border border-indigo-100 dark:border-slate-800"></pre>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3.5 border-t border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 flex items-center justify-between">
        <div class="text-[11px] text-slate-500 font-mono" id="modalMetadata">Latency: 0ms</div>
        <div class="flex items-center space-x-2">
          <button onclick="closeModal()" class="px-4 py-2 rounded-xl border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold transition shadow-2xs">Close</button>
          <button id="btnReplay" onclick="runReplay()" class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold transition flex items-center space-x-2 shadow-md shadow-indigo-500/20">
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
    let currentFilter = 'all';
    let latencyChartInstance = null;
    let toolPieChartInstance = null;

    function toggleTheme() {
      const html = document.documentElement;
      if (html.classList.contains('dark')) {
        html.classList.remove('dark');
        html.classList.add('light');
        document.getElementById('themeBtn').innerHTML = '<i class="fa-solid fa-moon text-xs"></i>';
        localStorage.setItem('theme', 'light');
      } else {
        html.classList.remove('light');
        html.classList.add('dark');
        document.getElementById('themeBtn').innerHTML = '<i class="fa-solid fa-sun text-xs"></i>';
        localStorage.setItem('theme', 'dark');
      }
      updateChartsTheme();
    }

    // Set initial theme
    if (localStorage.getItem('theme') === 'dark') {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    }

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

        sessions.forEach((s) => {
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
        renderCharts();
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
        document.getElementById('diagCommand').textContent = 'Subprocess: ' + curr.command;
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
        let badgeColor = 'bg-rose-50 dark:bg-rose-950/40 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-800/80';
        let icon = 'fa-repeat';
        if (f.failure_type === 'schema_mismatch') {
          badgeColor = 'bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-800/80';
          icon = 'fa-file-code';
        } else if (f.failure_type === 'slow_tool') {
          badgeColor = 'bg-orange-50 dark:bg-orange-950/40 text-orange-700 dark:text-orange-400 border-orange-200 dark:border-orange-800/80';
          icon = 'fa-hourglass-half';
        }

        div.className = 'p-3.5 rounded-xl border flex items-start space-x-3.5 shadow-2xs ' + badgeColor;
        div.innerHTML = '<i class="fa-solid ' + icon + ' mt-0.5 text-base"></i>' +
          '<div class="flex-1">' +
            '<div class="flex items-center justify-between">' +
              '<span class="font-bold text-xs uppercase tracking-wider">' + f.failure_type.replace('_', ' ') + '</span>' +
              '<span class="text-[10px] font-mono opacity-80">' + new Date(f.created_at).toLocaleTimeString() + '</span>' +
            '</div>' +
            '<p class="text-xs mt-1 font-medium">' + f.description + '</p>' +
          '</div>';
        list.appendChild(div);
      });
    }

    function setFilter(f) {
      currentFilter = f;
      ['All', 'Errors', 'Slow'].forEach(name => {
        const btn = document.getElementById('filter' + name);
        if (name.toLowerCase() === f) {
          btn.className = 'px-2.5 py-1 rounded-md bg-white dark:bg-slate-700 shadow-2xs text-slate-900 dark:text-white font-bold';
        } else {
          btn.className = 'px-2.5 py-1 rounded-md hover:text-slate-900 dark:hover:text-white font-medium';
        }
      });
      renderTraces();
    }

    function renderTraces() {
      const tbody = document.getElementById('traceTableBody');
      const search = document.getElementById('searchTraces').value.toLowerCase();
      document.getElementById('traceCountBadge').textContent = traces.length + ' traces';

      let filtered = traces.filter(t => t.tool_name.toLowerCase().includes(search));
      if (currentFilter === 'errors') {
        filtered = filtered.filter(t => t.is_error || t.status === 'failed');
      } else if (currentFilter === 'slow') {
        filtered = filtered.filter(t => t.latency_ms >= 1000);
      }

      if (filtered.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="py-12 text-center text-slate-400 font-sans">No matching tool calls found.</td></tr>';
        return;
      }

      tbody.innerHTML = '';
      filtered.forEach(t => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition cursor-pointer';

        let statusBadge = '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">SUCCESS</span>';
        if (t.is_error || t.status === 'failed') {
          statusBadge = '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-rose-100 dark:bg-rose-950/60 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-800">FAILED</span>';
        } else if (t.status === 'pending') {
          statusBadge = '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-100 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800 animate-pulse">RUNNING</span>';
        }

        let latColor = 'text-emerald-600 dark:text-emerald-400';
        let barWidth = Math.min(100, Math.max(8, (t.latency_ms / 3000) * 100));
        let barColor = 'bg-emerald-500';
        if (t.latency_ms > 1000) {
          latColor = 'text-amber-600 dark:text-amber-400';
          barColor = 'bg-amber-500';
        }
        if (t.latency_ms >= 5000) {
          latColor = 'text-rose-600 dark:text-rose-400 font-bold';
          barColor = 'bg-rose-500';
        }

        let argsPreview = t.arguments;
        if (argsPreview && argsPreview.length > 45) {
          argsPreview = argsPreview.substring(0, 45) + '...';
        }

        tr.innerHTML = 
          '<td class="py-3 px-4">' + statusBadge + '</td>' +
          '<td class="py-3 px-4 font-bold text-slate-800 dark:text-slate-200">' + t.tool_name + '</td>' +
          '<td class="py-3 px-4 text-slate-500 dark:text-slate-400 truncate max-w-xs">' + argsPreview + '</td>' +
          '<td class="py-3 px-4">' +
            '<div class="flex items-center space-x-2">' +
              '<span class="' + latColor + ' font-bold">' + t.latency_ms + ' ms</span>' +
              '<div class="w-16 bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">' +
                '<div class="' + barColor + ' h-full rounded-full" style="width: ' + barWidth + '%"></div>' +
              '</div>' +
            '</div>' +
          '</td>' +
          '<td class="py-3 px-4 text-slate-400 text-[11px]">' + new Date(t.created_at).toLocaleTimeString() + '</td>' +
          '<td class="py-3 px-4 text-right space-x-2 font-sans">' +
            '<button onclick="inspectTrace(\'' + t.id + '\')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-[11px] font-semibold transition">Inspect</button>' +
          '</td>';

        tr.onclick = (e) => {
          if (e.target.tagName !== 'BUTTON') {
            inspectTrace(t.id);
          }
        };
        tbody.appendChild(tr);
      });
    }

    function renderCharts() {
      const isDark = document.documentElement.classList.contains('dark');
      const gridColor = isDark ? 'rgba(51, 65, 85, 0.4)' : 'rgba(226, 232, 240, 0.8)';
      const textColor = isDark ? '#94a3b8' : '#64748b';

      // 1. Latency Line Chart
      const orderedTraces = [...traces].reverse();
      const labels = orderedTraces.map((t, i) => '#' + (i + 1) + ' ' + t.tool_name);
      const latencies = orderedTraces.map(t => t.latency_ms);

      const ctx1 = document.getElementById('latencyChart').getContext('2d');
      if (latencyChartInstance) latencyChartInstance.destroy();

      latencyChartInstance = new Chart(ctx1, {
        type: 'line',
        data: {
          labels: labels,
          datasets: [{
            label: 'Latency (ms)',
            data: latencies,
            borderColor: '#6366f1',
            backgroundColor: 'rgba(99, 102, 241, 0.1)',
            borderWidth: 2.5,
            fill: true,
            tension: 0.35,
            pointRadius: 4,
            pointBackgroundColor: '#6366f1'
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: { display: false }
          },
          scales: {
            x: {
              grid: { color: gridColor },
              ticks: { color: textColor, font: { size: 10 } }
            },
            y: {
              grid: { color: gridColor },
              ticks: { color: textColor, font: { size: 10 } },
              beginAtZero: true
            }
          }
        }
      });

      // 2. Tool Distribution Pie Chart
      const toolCounts = {};
      traces.forEach(t => {
        toolCounts[t.tool_name] = (toolCounts[t.tool_name] || 0) + 1;
      });

      const pieLabels = Object.keys(toolCounts);
      const pieData = Object.values(toolCounts);

      const ctx2 = document.getElementById('toolPieChart').getContext('2d');
      if (toolPieChartInstance) toolPieChartInstance.destroy();

      toolPieChartInstance = new Chart(ctx2, {
        type: 'doughnut',
        data: {
          labels: pieLabels,
          datasets: [{
            data: pieData,
            backgroundColor: ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6'],
            borderWidth: 2,
            borderColor: isDark ? '#0f172a' : '#ffffff'
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              position: 'bottom',
              labels: { color: textColor, boxWidth: 10, font: { size: 10 } }
            }
          },
          cutout: '65%'
        }
      });
    }

    function updateChartsTheme() {
      if (latencyChartInstance && toolPieChartInstance) {
        renderCharts();
      }
    }

    function inspectTrace(id) {
      const trace = traces.find(t => t.id === id);
      if (!trace) return;
      currentModalTrace = trace;

      document.getElementById('modalToolName').textContent = trace.tool_name;
      const statusBadge = document.getElementById('modalStatusBadge');
      if (trace.is_error) {
        statusBadge.textContent = 'Failed';
        statusBadge.className = 'text-[10px] font-bold px-2.5 py-0.5 rounded-full uppercase bg-rose-100 text-rose-700 border border-rose-200';
      } else {
        statusBadge.textContent = 'Success';
        statusBadge.className = 'text-[10px] font-bold px-2.5 py-0.5 rounded-full uppercase bg-emerald-100 text-emerald-700 border border-emerald-200';
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
