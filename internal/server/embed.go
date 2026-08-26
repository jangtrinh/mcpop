package server

import (
	_ "embed"
	"net/http"
)

// Embedded premium Neutral B&W Monochrome Dashboard
const DashboardHTML = `<!DOCTYPE html>
<html lang="en" class="light">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MCPOp — MCP Observability & Failure Catcher</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
  <!-- Phosphor Icons -->
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
  <link href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:ital,wght@0,400;0,500;0,600;0,700;0,800;1,400;1,600&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    body { font-family: 'Be Vietnam Pro', sans-serif; font-size: 13px; letter-spacing: -0.01em; }
    code, pre, .font-mono { font-family: 'JetBrains Mono', monospace; }
    ::-webkit-scrollbar { width: 5px; height: 5px; }
    ::-webkit-scrollbar-track { background: transparent; }
    ::-webkit-scrollbar-thumb { background: #d4d4d8; border-radius: 9999px; }
    .dark ::-webkit-scrollbar-thumb { background: #3f3f46; }
  </style>
</head>
<body class="bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100 min-h-screen flex flex-col antialiased transition-colors duration-150 selection:bg-neutral-900 selection:text-white dark:selection:bg-white dark:selection:text-black">

  <!-- Top Navigation Bar (High-end Neutral B&W) -->
  <header class="border-b border-neutral-200/90 dark:border-neutral-800/90 bg-white/95 dark:bg-neutral-900/95 backdrop-blur sticky top-0 z-30 px-6 py-3.5 flex items-center justify-between shadow-2xs">
    <div class="flex items-center space-x-3.5">
      <div class="w-8 h-8 rounded-lg bg-neutral-950 dark:bg-white text-white dark:text-neutral-950 flex items-center justify-center font-bold shadow-xs" aria-hidden="true">
        <i class="ph-bold ph-terminal text-[16px]"></i>
      </div>
      <div>
        <div class="flex items-center space-x-2">
          <span class="font-extrabold tracking-tight text-neutral-950 dark:text-white text-base">MCPOp</span>
          <span class="text-[11px] font-semibold bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 border border-neutral-200 dark:border-neutral-700 px-2 py-0.5 rounded-md uppercase tracking-wider">Observability</span>
          <span class="text-[11px] font-mono font-medium text-neutral-500 bg-neutral-100 dark:bg-neutral-800 px-1.5 py-0.5 rounded">v0.1.0</span>
        </div>
        <p class="text-xs text-neutral-500 dark:text-neutral-400 font-medium">Silent Failure Catcher & Realtime Waterfall for MCP</p>
      </div>
    </div>

    <!-- Active Session, Live Indicator & Theme Toggle -->
    <div class="flex items-center space-x-3">
      <!-- Live Indicator -->
      <div class="flex items-center space-x-2 text-xs bg-neutral-100 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 px-3 py-1.5 rounded-lg text-neutral-800 dark:text-neutral-200 font-medium">
        <span class="relative flex h-2 w-2" aria-hidden="true">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-neutral-900 dark:bg-white opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-neutral-900 dark:bg-white"></span>
        </span>
        <span id="liveStatus">Live SSE Connected</span>
      </div>

      <!-- Session Picker -->
      <div class="flex items-center space-x-1 bg-white dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-700 rounded-lg p-1 shadow-2xs">
        <label for="sessionSelect" class="sr-only">Select Session</label>
        <i class="ph ph-terminal-window text-[16px] text-neutral-400 pl-2" aria-hidden="true"></i>
        <select id="sessionSelect" aria-label="Target MCP Session" class="bg-transparent text-xs text-neutral-800 dark:text-neutral-200 font-medium py-1.5 px-2 focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white rounded cursor-pointer min-h-[32px]">
          <option value="">Loading sessions...</option>
        </select>
        <button onclick="refreshData()" aria-label="Refresh Data" title="Refresh" class="w-8 h-8 flex items-center justify-center hover:bg-neutral-100 dark:hover:bg-neutral-700 rounded text-neutral-500 hover:text-neutral-900 dark:hover:text-white transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">
          <i class="ph ph-arrows-clockwise text-[16px]"></i>
        </button>
      </div>

      <!-- Theme Switcher -->
      <button onclick="toggleTheme()" id="themeBtn" aria-label="Toggle Dark/Light Theme" class="w-8 h-8 rounded-lg border border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-neutral-700 dark:text-neutral-200 hover:bg-neutral-100 dark:hover:bg-neutral-700 flex items-center justify-center transition shadow-2xs focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">
        <i class="ph ph-moon text-[16px]"></i>
      </button>
    </div>
  </header>

  <!-- Main Container -->
  <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">

    <!-- Interactive Architecture Flow Diagram (Neutral B&W) -->
    <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-5 shadow-2xs">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center space-x-2 text-xs font-bold text-neutral-900 dark:text-white uppercase tracking-wider">
          <i class="ph-bold ph-network text-[18px]" aria-hidden="true"></i>
          <span>Transparent Interceptor Topology</span>
        </div>
        <span class="text-[11px] text-neutral-500 font-mono" id="diagCommand">Command: (waiting for session)</span>
      </div>

      <!-- Visual Flow Chart Diagram -->
      <div class="grid grid-cols-1 md:grid-cols-5 gap-3 items-center text-center">
        <!-- Node 1: AI Client -->
        <div class="bg-neutral-50 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 rounded-lg p-3.5">
          <div class="w-8 h-8 mx-auto rounded-md bg-neutral-200 dark:bg-neutral-700 text-neutral-900 dark:text-white flex items-center justify-center mb-1.5" aria-hidden="true">
            <i class="ph ph-robot text-[18px]"></i>
          </div>
          <div class="text-xs font-bold text-neutral-900 dark:text-white">AI Client</div>
          <div class="text-[11px] text-neutral-500">Claude / Cursor / Agent</div>
        </div>

        <!-- Connection Arrow 1 -->
        <div class="flex flex-col items-center justify-center text-neutral-400 py-1" aria-hidden="true">
          <span class="text-[11px] font-mono font-medium text-neutral-700 dark:text-neutral-300 bg-neutral-100 dark:bg-neutral-800 px-2.5 py-0.5 rounded border border-neutral-200 dark:border-neutral-700 mb-1">stdio pipe</span>
          <i class="ph ph-arrows-left-right text-[16px] hidden md:block"></i>
          <i class="ph ph-arrows-down-up text-[16px] md:hidden"></i>
        </div>

        <!-- Node 2: MCPOp Engine (Hero Neutral Focus) -->
        <div class="bg-neutral-900 dark:bg-neutral-100 text-white dark:text-neutral-950 border border-neutral-950 dark:border-white rounded-lg p-3.5 relative shadow-xs">
          <div class="absolute -top-2.5 left-1/2 -translate-x-1/2 bg-neutral-950 dark:bg-white text-white dark:text-neutral-950 text-[11px] font-mono font-bold uppercase px-2.5 py-0.5 rounded border border-neutral-700 dark:border-neutral-300 tracking-wider">
            Active Proxy
          </div>
          <div class="w-8 h-8 mx-auto rounded-md bg-neutral-800 dark:bg-neutral-200 flex items-center justify-center mb-1.5" aria-hidden="true">
            <i class="ph ph-shield-check text-[18px]"></i>
          </div>
          <div class="text-xs font-bold">MCPOp Core</div>
          <div class="text-[11px] opacity-80 font-mono font-medium">&lt;1ms Interceptor</div>
        </div>

        <!-- Connection Arrow 2 -->
        <div class="flex flex-col items-center justify-center text-neutral-400 py-1" aria-hidden="true">
          <span class="text-[11px] font-mono font-medium text-neutral-700 dark:text-neutral-300 bg-neutral-100 dark:bg-neutral-800 px-2.5 py-0.5 rounded border border-neutral-200 dark:border-neutral-700 mb-1">JSON-RPC 2.0</span>
          <i class="ph ph-arrows-left-right text-[16px] hidden md:block"></i>
          <i class="ph ph-arrows-down-up text-[16px] md:hidden"></i>
        </div>

        <!-- Node 3: Target Server -->
        <div class="bg-neutral-50 dark:bg-neutral-800/80 border border-neutral-200 dark:border-neutral-700 rounded-lg p-3.5">
          <div class="w-8 h-8 mx-auto rounded-md bg-neutral-200 dark:bg-neutral-700 text-neutral-900 dark:text-white flex items-center justify-center mb-1.5" aria-hidden="true">
            <i class="ph ph-hard-drives text-[18px]"></i>
          </div>
          <div class="text-xs font-bold text-neutral-900 dark:text-white">Target Server</div>
          <div class="text-[11px] text-neutral-500">Python / Node / Go Tools</div>
        </div>
      </div>
    </div>

    <!-- Stats KPI Cards (Clean B&W Monochrome) -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-4 shadow-2xs hover:border-neutral-400 dark:hover:border-neutral-600 transition">
        <div class="flex items-center justify-between text-neutral-500 text-xs font-semibold">
          <span>Total Tool Invocations</span>
          <div class="w-7 h-7 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 flex items-center justify-center" aria-hidden="true">
            <i class="ph ph-stack text-[16px]"></i>
          </div>
        </div>
        <div class="text-2xl font-extrabold text-neutral-950 dark:text-white mt-2" id="statTotalCalls">0</div>
        <div class="text-[11px] text-neutral-500 mt-1 flex items-center space-x-1.5 font-medium">
          <i class="ph ph-chart-bar text-[16px]" aria-hidden="true"></i>
          <span>Captured in real-time</span>
        </div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-4 shadow-2xs hover:border-neutral-400 dark:hover:border-neutral-600 transition">
        <div class="flex items-center justify-between text-neutral-500 text-xs font-semibold">
          <span>Success Reliability</span>
          <div class="w-7 h-7 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 flex items-center justify-center" aria-hidden="true">
            <i class="ph ph-check-circle text-[16px]"></i>
          </div>
        </div>
        <div class="text-2xl font-extrabold text-neutral-950 dark:text-white mt-2" id="statSuccessRate">100%</div>
        <div class="text-[11px] text-neutral-500 mt-1 flex items-center space-x-1.5 font-medium" id="statErrorCallsBox">
          <span id="statErrorCalls">0 errors</span>
        </div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-4 shadow-2xs hover:border-neutral-400 dark:hover:border-neutral-600 transition">
        <div class="flex items-center justify-between text-neutral-500 text-xs font-semibold">
          <span>Average Latency</span>
          <div class="w-7 h-7 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 flex items-center justify-center" aria-hidden="true">
            <i class="ph ph-timer text-[16px]"></i>
          </div>
        </div>
        <div class="text-2xl font-extrabold text-neutral-950 dark:text-white mt-2" id="statAvgLatency">0 ms</div>
        <div class="text-[11px] text-neutral-500 mt-1 flex items-center space-x-1 font-medium">
          <span>Execution time per call</span>
        </div>
      </div>

      <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-4 shadow-2xs hover:border-neutral-400 dark:hover:border-neutral-600 transition">
        <div class="flex items-center justify-between text-neutral-500 text-xs font-semibold">
          <span>Silent Failures Intercepted</span>
          <div class="w-7 h-7 rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-800 dark:text-neutral-200 flex items-center justify-center" aria-hidden="true">
            <i class="ph ph-warning text-[16px]"></i>
          </div>
        </div>
        <div class="text-2xl font-extrabold text-neutral-950 dark:text-white mt-2" id="statFailures">0</div>
        <div class="text-[11px] text-neutral-500 mt-1 flex items-center space-x-1 font-medium">
          <span>Loops, Schemas & Slow Tools</span>
        </div>
      </div>
    </div>

    <!-- Charts Section (Monochrome Analytics) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Chart 1: Latency Timeline (2 cols) -->
      <div class="lg:col-span-2 bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-5 shadow-2xs">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-xs font-bold text-neutral-900 dark:text-white uppercase tracking-wider flex items-center space-x-2">
              <i class="ph-bold ph-chart-line-up text-[18px]" aria-hidden="true"></i>
              <span>Tool Call Latency Timeline (ms)</span>
            </h3>
            <p class="text-[11px] text-neutral-500 font-medium">Execution latency progression across sequence</p>
          </div>
          <span class="text-[11px] font-mono bg-neutral-100 dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 px-2 py-0.5 rounded font-semibold">Realtime</span>
        </div>
        <div class="relative h-48 w-full">
          <canvas id="latencyChart" role="img" aria-label="Tool Latency Timeline Chart"></canvas>
        </div>
      </div>

      <!-- Chart 2: Tool Shares Doughnut (1 col) -->
      <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl p-5 shadow-2xs flex flex-col justify-between">
        <div>
          <h3 class="text-xs font-bold text-neutral-900 dark:text-white uppercase tracking-wider flex items-center space-x-2 mb-1">
            <i class="ph-bold ph-chart-pie-slice text-[18px]" aria-hidden="true"></i>
            <span>Tool Invocation Shares</span>
          </h3>
          <p class="text-[11px] text-neutral-500 font-medium mb-3">Volume breakdown by tool name</p>
        </div>
        <div class="relative h-44 flex items-center justify-center">
          <canvas id="toolPieChart" role="img" aria-label="Tool Invocation Share Doughnut Chart"></canvas>
        </div>
      </div>
    </div>

    <!-- Failure Alerts Section -->
    <div id="failureAlertsContainer" class="hidden space-y-2.5" role="region" aria-label="Failure Alerts">
      <div class="flex items-center space-x-2 text-xs font-bold uppercase tracking-wider text-neutral-950 dark:text-white">
        <i class="ph-bold ph-warning-circle text-[18px]" aria-hidden="true"></i>
        <span>Heuristic Anomaly Detections</span>
      </div>
      <div id="failureAlertsList" class="space-y-2"></div>
    </div>

    <!-- Traces Waterfall Table -->
    <div class="bg-white dark:bg-neutral-900 border border-neutral-200/90 dark:border-neutral-800/90 rounded-xl overflow-hidden shadow-2xs">
      <div class="p-4 border-b border-neutral-200/90 dark:border-neutral-800/90 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center space-x-3">
          <h2 class="text-sm font-bold text-neutral-950 dark:text-white flex items-center space-x-2">
            <span>Live Tool Trace Waterfall</span>
          </h2>
          <span id="traceCountBadge" class="text-[11px] bg-neutral-100 dark:bg-neutral-800 text-neutral-700 dark:text-neutral-300 px-2.5 py-0.5 rounded font-mono font-semibold">0 traces</span>
        </div>

        <div class="flex items-center space-x-2.5">
          <!-- Filter Buttons -->
          <div class="flex items-center bg-neutral-100 dark:bg-neutral-800 p-0.5 rounded-lg text-xs font-medium text-neutral-600 dark:text-neutral-400" role="group" aria-label="Filter traces">
            <button onclick="setFilter('all')" id="filterAll" class="px-3 py-1.5 rounded-md bg-white dark:bg-neutral-700 shadow-2xs text-neutral-950 dark:text-white font-bold transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">All</button>
            <button onclick="setFilter('errors')" id="filterErrors" class="px-3 py-1.5 rounded-md hover:text-neutral-950 dark:hover:text-white transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">Errors</button>
            <button onclick="setFilter('slow')" id="filterSlow" class="px-3 py-1.5 rounded-md hover:text-neutral-950 dark:hover:text-white transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">Slow</button>
          </div>

          <div class="relative">
            <label for="searchTraces" class="sr-only">Search tool traces</label>
            <i class="ph ph-magnifying-glass absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400 text-[16px]" aria-hidden="true"></i>
            <input id="searchTraces" oninput="renderTraces()" type="text" placeholder="Search tools..." aria-label="Search tools by name" class="bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-700 text-xs rounded-lg pl-8 pr-3 py-1.5 text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 dark:placeholder-neutral-500 focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white focus:border-neutral-950 w-44 min-h-[32px]">
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs" role="table">
          <thead class="bg-neutral-50 dark:bg-neutral-950/80 text-neutral-600 dark:text-neutral-400 border-b border-neutral-200/90 dark:border-neutral-800/90 font-semibold">
            <tr>
              <th scope="col" class="py-3 px-4">Status</th>
              <th scope="col" class="py-3 px-4">Tool Name</th>
              <th scope="col" class="py-3 px-4">Arguments Payload</th>
              <th scope="col" class="py-3 px-4">Latency</th>
              <th scope="col" class="py-3 px-4">Timestamp</th>
              <th scope="col" class="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody id="traceTableBody" class="divide-y divide-neutral-200/60 dark:divide-neutral-800/60 font-mono">
            <tr>
              <td colspan="6" class="py-12 text-center text-neutral-400 font-sans">No tool calls recorded in this session yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </main>

  <!-- Tool Detail & Replay Drawer Modal (Neutral B&W) -->
  <div id="detailModal" onclick="handleBackdropClick(event)" class="fixed inset-0 bg-neutral-950/60 backdrop-blur-xs z-50 hidden flex items-center justify-center p-4" role="dialog" aria-modal="true" aria-labelledby="modalToolName">
    <div id="modalCard" class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-xl max-w-3xl w-full shadow-2xl flex flex-col max-h-[90vh] overflow-hidden">
      
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-neutral-200 dark:border-neutral-800 flex items-center justify-between bg-neutral-50/50 dark:bg-neutral-900">
        <div class="flex items-center space-x-2.5">
          <div class="w-2.5 h-2.5 rounded-full bg-neutral-950 dark:bg-white" aria-hidden="true"></div>
          <h3 class="font-bold text-neutral-950 dark:text-white text-base font-mono" id="modalToolName">tool/name</h3>
          <span id="modalStatusBadge" class="text-[11px] font-bold px-2.5 py-0.5 rounded uppercase"></span>
        </div>
        <button onclick="closeModal()" aria-label="Close dialog" class="w-8 h-8 rounded-lg flex items-center justify-center text-neutral-400 hover:text-neutral-950 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-neutral-800 transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">
          <i class="ph ph-x text-[18px]"></i>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 overflow-y-auto space-y-5 text-xs">
        <div>
          <div class="flex items-center justify-between mb-1.5">
            <label for="modalArgsInput" class="block text-neutral-700 dark:text-neutral-300 font-bold uppercase tracking-wider text-[11px]">Input Arguments (JSON)</label>
            <button onclick="copyArgsJSON()" id="btnCopyArgs" class="text-[11px] text-neutral-900 dark:text-neutral-100 hover:underline flex items-center space-x-1 font-sans font-semibold">
              <i class="ph ph-copy text-[16px]"></i>
              <span id="copyArgsText">Copy JSON</span>
            </button>
          </div>
          <textarea id="modalArgsInput" aria-label="Input arguments in JSON" class="w-full bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 font-mono text-neutral-900 dark:text-neutral-100 text-xs focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white h-32"></textarea>
        </div>

        <div>
          <label class="block text-neutral-700 dark:text-neutral-300 font-bold mb-1.5 uppercase tracking-wider text-[11px]">Execution Result Payload</label>
          <pre id="modalResultText" class="w-full bg-neutral-50 dark:bg-neutral-950 border border-neutral-200 dark:border-neutral-800 rounded-lg p-3 font-mono text-neutral-800 dark:text-neutral-200 text-xs overflow-x-auto max-h-48 whitespace-pre-wrap"></pre>
        </div>

        <!-- Replay Result Box -->
        <div id="replayResultBox" class="hidden p-4 rounded-lg border border-neutral-300 dark:border-neutral-700 bg-neutral-100/70 dark:bg-neutral-800/40 space-y-2">
          <div class="flex items-center justify-between">
            <span class="font-bold text-neutral-950 dark:text-white text-xs">Replay Execution Response</span>
            <span id="replayLatencyBadge" class="text-[11px] font-mono font-bold text-neutral-700 dark:text-neutral-300"></span>
          </div>
          <pre id="replayOutputText" class="font-mono text-xs text-neutral-900 dark:text-neutral-100 whitespace-pre-wrap bg-white dark:bg-neutral-950 p-3 rounded border border-neutral-200 dark:border-neutral-700"></pre>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3.5 border-t border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-950 flex items-center justify-between">
        <div class="text-[11px] text-neutral-500 font-mono" id="modalMetadata">Latency: 0ms</div>
        <div class="flex items-center space-x-2">
          <button onclick="closeModal()" class="px-4 py-2 rounded-lg border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 hover:bg-neutral-100 dark:hover:bg-neutral-700 text-neutral-800 dark:text-neutral-200 text-xs font-semibold transition shadow-2xs focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">Close</button>
          <button id="btnReplay" onclick="runReplay()" class="px-4 py-2 rounded-lg bg-neutral-950 hover:bg-neutral-800 dark:bg-white dark:hover:bg-neutral-200 text-white dark:text-neutral-950 text-xs font-bold transition flex items-center space-x-2 shadow-xs focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">
            <i class="ph ph-play text-[16px]"></i>
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
        document.getElementById('themeBtn').innerHTML = '<i class="ph ph-moon text-[16px]"></i>';
        localStorage.setItem('theme', 'light');
      } else {
        html.classList.remove('light');
        html.classList.add('dark');
        document.getElementById('themeBtn').innerHTML = '<i class="ph ph-sun text-[16px]"></i>';
        localStorage.setItem('theme', 'dark');
      }
      updateChartsTheme();
    }

    if (localStorage.getItem('theme') === 'dark') {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    }

    async function init() {
      await fetchSessions();
      setupSSE();

      window.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
          closeModal();
        }
      });
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
        let icon = 'ph-repeat';
        if (f.failure_type === 'schema_mismatch') icon = 'ph-code';
        else if (f.failure_type === 'slow_tool') icon = 'ph-hourglass';

        div.className = 'p-3.5 rounded-lg border border-neutral-300 dark:border-neutral-700 bg-neutral-100/90 dark:bg-neutral-800/80 text-neutral-900 dark:text-neutral-100 flex items-start space-x-3.5 shadow-2xs';
        div.innerHTML = '<i class="ph-bold ' + icon + ' mt-0.5 text-[18px]"></i>' +
          '<div class="flex-1">' +
            '<div class="flex items-center justify-between">' +
              '<span class="font-bold text-xs uppercase tracking-wider">' + f.failure_type.replace('_', ' ') + '</span>' +
              '<span class="text-[11px] font-mono opacity-60">' + new Date(f.created_at).toLocaleTimeString() + '</span>' +
            '</div>' +
            '<p class="text-xs mt-1 font-medium text-neutral-700 dark:text-neutral-300">' + f.description + '</p>' +
          '</div>';
        list.appendChild(div);
      });
    }

    function setFilter(f) {
      currentFilter = f;
      ['All', 'Errors', 'Slow'].forEach(name => {
        const btn = document.getElementById('filter' + name);
        if (name.toLowerCase() === f) {
          btn.className = 'px-3 py-1.5 rounded-md bg-white dark:bg-neutral-700 shadow-2xs text-neutral-950 dark:text-white font-bold transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white';
        } else {
          btn.className = 'px-3 py-1.5 rounded-md hover:text-neutral-950 dark:hover:text-white transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white font-medium';
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
        tbody.innerHTML = '<tr><td colspan="6" class="py-12 text-center text-neutral-400 font-sans">No matching tool calls found.</td></tr>';
        return;
      }

      tbody.innerHTML = '';
      filtered.forEach(t => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-neutral-100/60 dark:hover:bg-neutral-800/40 transition cursor-pointer';

        let statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-neutral-200 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 border border-neutral-300 dark:border-neutral-700">OK</span>';
        if (t.is_error || t.status === 'failed') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-neutral-950 text-white dark:bg-white dark:text-black">FAIL</span>';
        } else if (t.status === 'pending') {
          statusBadge = '<span class="px-2 py-0.5 rounded text-[11px] font-bold bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-400 border border-neutral-300 dark:border-neutral-700 animate-pulse">RUN</span>';
        }

        let argsPreview = t.arguments;
        if (argsPreview && argsPreview.length > 45) {
          argsPreview = argsPreview.substring(0, 45) + '...';
        }

        tr.innerHTML = 
          '<td class="py-3 px-4">' + statusBadge + '</td>' +
          '<td class="py-3 px-4 font-bold text-neutral-950 dark:text-white">' + t.tool_name + '</td>' +
          '<td class="py-3 px-4 text-neutral-500 truncate max-w-xs">' + argsPreview + '</td>' +
          '<td class="py-3 px-4 font-bold font-mono text-neutral-900 dark:text-neutral-100">' + t.latency_ms + ' ms</td>' +
          '<td class="py-3 px-4 text-neutral-400 text-[11px]">' + new Date(t.created_at).toLocaleTimeString() + '</td>' +
          '<td class="py-3 px-4 text-right space-x-2 font-sans">' +
            '<button onclick="inspectTrace(\'' + t.id + '\')" aria-label="Inspect tool call ' + t.tool_name + '" class="px-3 py-1.5 rounded-md border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-800 hover:bg-neutral-100 dark:hover:bg-neutral-700 text-neutral-900 dark:text-neutral-100 text-xs font-semibold transition focus:outline-none focus:ring-1 focus:ring-neutral-950 dark:focus:ring-white">Inspect</button>' +
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
      const gridColor = isDark ? 'rgba(63, 63, 70, 0.4)' : 'rgba(228, 228, 231, 0.8)';
      const textColor = isDark ? '#a1a1aa' : '#71717a';
      const lineColor = isDark ? '#ffffff' : '#18181b';

      // 1. Latency Line Chart (Monochrome)
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
            borderColor: lineColor,
            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(24, 24, 27, 0.04)',
            borderWidth: 2,
            fill: true,
            tension: 0.25,
            pointRadius: 3,
            pointBackgroundColor: lineColor
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
              ticks: { color: textColor, font: { family: 'Be Vietnam Pro', size: 11 } }
            },
            y: {
              grid: { color: gridColor },
              ticks: { color: textColor, font: { family: 'Be Vietnam Pro', size: 11 } },
              beginAtZero: true
            }
          }
        }
      });

      // 2. Tool Distribution Pie Chart (Monochrome Shades)
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
              position: 'bottom',
              labels: { color: textColor, boxWidth: 10, font: { family: 'Be Vietnam Pro', size: 11 } }
            }
          },
          cutout: '70%'
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
        statusBadge.className = 'text-[11px] font-bold px-2.5 py-0.5 rounded uppercase bg-neutral-950 text-white dark:bg-white dark:text-black';
      } else {
        statusBadge.textContent = 'Success';
        statusBadge.className = 'text-[11px] font-bold px-2.5 py-0.5 rounded uppercase bg-neutral-200 text-neutral-900 dark:bg-neutral-800 dark:text-neutral-100 border border-neutral-300 dark:border-neutral-700';
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
      if (e.target.id === 'detailModal') {
        closeModal();
      }
    }

    function copyArgsJSON() {
      const text = document.getElementById('modalArgsInput').value;
      navigator.clipboard.writeText(text).then(() => {
        const label = document.getElementById('copyArgsText');
        label.textContent = 'Copied!';
        setTimeout(() => {
          label.textContent = 'Copy JSON';
        }, 2000);
      });
    }

    async function runReplay() {
      if (!currentModalTrace) return;
      const btn = document.getElementById('btnReplay');
      const box = document.getElementById('replayResultBox');
      const out = document.getElementById('replayOutputText');
      const lat = document.getElementById('replayLatencyBadge');

      btn.disabled = true;
      btn.innerHTML = '<i class="ph ph-spinner animate-spin text-[16px]"></i> <span>Replaying...</span>';

      const session = sessions.find(s => s.id === currentSessionId);
      const command = session ? session.command : '';

      let parsedArgs = {};
      try {
        parsedArgs = JSON.parse(document.getElementById('modalArgsInput').value);
      } catch (e) {
        alert('Invalid JSON in arguments');
        btn.disabled = false;
        btn.innerHTML = '<i class="ph ph-play text-[16px]"></i> <span>1-Click Replay</span>';
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
        btn.innerHTML = '<i class="ph ph-play text-[16px]"></i> <span>1-Click Replay</span>';
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
