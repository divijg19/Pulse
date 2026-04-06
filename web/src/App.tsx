import { createSignal } from "solid-js";

function App() {
	// SolidJS Signal to control our tabs
	const [activeTab, setActiveTab] = createSignal("timeline");

	return (
		// ROOT APP CONTAINER: Locked to 100vh, no page scroll, ambient background
		<div class="h-screen w-screen overflow-hidden bg-slate-950 flex flex-col font-sans text-slate-300 relative">
			{/* Ambient Background Glow */}
			<div class="absolute inset-0 z-0 bg-[radial-gradient(ellipse_at_top,var(--tw-gradient-stops))] from-teal-900/20 via-slate-950 to-slate-950 pointer-events-none"></div>

			{/* MAIN CONTENT WRAPPER: relative z-10 puts it above the background */}
			<div class="relative z-10 flex flex-col h-full p-6 max-w-7xl mx-auto w-full gap-6">
				{/* --- HEADER --- */}
				<header class="flex justify-between items-end shrink-0">
					<div class="flex items-center gap-3">
						<div class="relative flex h-4 w-4">
							<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-teal-400 opacity-75"></span>
							<span class="relative inline-flex rounded-full h-4 w-4 bg-teal-500 shadow-[0_0_10px_rgba(20,184,166,0.8)]"></span>
						</div>
						<h1 class="text-3xl font-extrabold tracking-tight text-white">
							Pulse
						</h1>
					</div>
					<span class="text-slate-500 text-sm font-mono tracking-widest">
						HTTP_OBSERVABILITY_ENGINE
					</span>
				</header>

				{/* --- THE UNIFIED CONTROL PILL --- */}
				<div class="shrink-0 bg-slate-900/50 backdrop-blur-md border border-slate-700/50 rounded-2xl p-2 shadow-2xl flex items-center transition-all hover:border-slate-600/50 focus-within:border-teal-500/50 focus-within:shadow-[0_0_30px_rgba(20,184,166,0.15)]">
					<select class="bg-transparent border-none text-teal-400 font-bold px-6 py-3 outline-none cursor-pointer appearance-none uppercase tracking-wider">
						<option class="bg-slate-900 text-teal-400">GET</option>
						<option class="bg-slate-900 text-blue-400">POST</option>
						<option class="bg-slate-900 text-amber-400">PUT</option>
						<option class="bg-slate-900 text-rose-400">DELETE</option>
					</select>

					<div class="w-px h-8 bg-slate-700/50 mx-2"></div>

					<input
						type="text"
						value="https://httpbin.org/delay/1"
						class="flex-1 bg-transparent border-none px-4 py-3 outline-none font-mono text-slate-200 placeholder-slate-600 text-lg"
						placeholder="Enter API endpoint..."
					/>

					<div class="w-px h-8 bg-slate-700/50 mx-2"></div>

					<div class="flex items-center gap-2 px-6">
						<svg
							class="w-5 h-5 text-slate-500"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							aria-label="Request frequency"
						>
							<title>Request frequency</title>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 10V3L4 14h7v7l9-11h-7z"
							></path>
						</svg>
						<input
							type="number"
							value="10"
							class="bg-transparent w-12 outline-none text-white font-mono text-lg text-center"
						/>
					</div>

					<button
						type="button"
						class="bg-linear-to-r from-teal-500 to-emerald-400 hover:from-teal-400 hover:to-emerald-300 text-slate-950 font-black px-10 py-3 rounded-xl shadow-[0_0_20px_rgba(20,184,166,0.4)] transition-all active:scale-95 uppercase tracking-widest ml-2"
					>
						Fire
					</button>
				</div>

				{/* --- METRICS STRIP --- */}
				<div class="shrink-0 grid grid-cols-4 gap-4">
					{[
						{ label: "Requests", val: "100", color: "text-white" },
						{ label: "Success Rate", val: "100%", color: "text-teal-400" },
						{ label: "Avg Latency", val: "1.24s", color: "text-amber-400" },
						{ label: "Throughput", val: "8.5 rps", color: "text-fuchsia-400" },
					].map((m) => (
						<div class="bg-slate-900/40 backdrop-blur border border-slate-800/50 rounded-xl p-4 flex flex-col justify-center items-center relative overflow-hidden group hover:bg-slate-800/50 transition-colors">
							<span class="text-slate-500 text-xs font-bold uppercase tracking-wider mb-1">
								{m.label}
							</span>
							<span class={`text-2xl font-mono ${m.color}`}>{m.val}</span>
						</div>
					))}
				</div>

				{/* --- TABBED WORKSPACE (The Brain) --- */}
				{/* flex-1 allows this container to fill remaining height. min-h-0 prevents it from pushing past 100vh */}
				<div class="flex-1 min-h-0 flex flex-col bg-slate-900/60 backdrop-blur-md border border-slate-700/50 rounded-2xl overflow-hidden shadow-2xl">
					{/* Tab Headers */}
					<div class="flex border-b border-slate-700/50 bg-slate-950/30">
						<button
							type="button"
							onClick={() => setActiveTab("timeline")}
							class={`flex-1 py-4 text-sm font-bold uppercase tracking-widest transition-colors ${activeTab() === "timeline" ? "text-teal-400 border-b-2 border-teal-400 bg-slate-800/30" : "text-slate-500 hover:text-slate-300"}`}
						>
							📊 Performance Timeline
						</button>
						<button
							type="button"
							onClick={() => setActiveTab("logs")}
							class={`flex-1 py-4 text-sm font-bold uppercase tracking-widest transition-colors ${activeTab() === "logs" ? "text-teal-400 border-b-2 border-teal-400 bg-slate-800/30" : "text-slate-500 hover:text-slate-300"}`}
						>
							⌨️ Live Terminal Logs
						</button>
					</div>

					{/* Tab Content Area: internally scrollable */}
					<div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
						{/* TIMELINE VIEW */}
						{activeTab() === "timeline" && (
							<div class="space-y-3">
								{/* Dummy Bar 1 */}
								<div class="w-full bg-slate-950/50 rounded-lg h-10 relative overflow-hidden flex items-center px-4 border border-slate-800/50 group hover:border-slate-700">
									<div class="absolute left-0 top-0 bottom-0 bg-linear-to-r from-teal-500/10 to-teal-500/30 w-[60%] border-r-2 border-teal-400 shadow-[2px_0_10px_rgba(20,184,166,0.3)]"></div>
									<span class="relative z-10 text-sm font-mono text-slate-300 group-hover:text-white transition-colors">
										200 OK
									</span>
									<span class="relative z-10 text-sm font-mono text-teal-400 ml-auto">
										1.24s
									</span>
								</div>
								{/* Dummy Bar 2 */}
								<div class="w-full bg-slate-950/50 rounded-lg h-10 relative overflow-hidden flex items-center px-4 border border-slate-800/50 group hover:border-slate-700">
									<div class="absolute left-0 top-0 bottom-0 bg-linear-to-r from-rose-500/10 to-rose-500/30 w-[95%] border-r-2 border-rose-400 shadow-[2px_0_10px_rgba(244,63,94,0.3)]"></div>
									<span class="relative z-10 text-sm font-mono text-slate-300 group-hover:text-white transition-colors">
										500 Internal Error
									</span>
									<span class="relative z-10 text-sm font-mono text-rose-400 ml-auto">
										3.10s
									</span>
								</div>
							</div>
						)}

						{/* LOGS VIEW */}
						{activeTab() === "logs" && (
							<div class="font-mono text-sm space-y-2">
								<div class="flex gap-4 p-2 rounded hover:bg-slate-800/50 transition-colors">
									<span class="text-slate-600">12:01:00.145</span>
									<span class="text-teal-400 font-bold">GET</span>
									<span class="text-slate-300 truncate flex-1">
										https://httpbin.org/delay/1
									</span>
									<span class="text-teal-400">200 OK</span>
									<span class="text-slate-500">1.24s</span>
								</div>
								<div class="flex gap-4 p-2 rounded bg-rose-950/20 border border-rose-900/30 hover:bg-rose-900/20 transition-colors">
									<span class="text-slate-600">12:01:01.320</span>
									<span class="text-rose-400 font-bold">GET</span>
									<span class="text-slate-300 truncate flex-1">
										https://httpbin.org/delay/1
									</span>
									<span class="text-rose-500 font-bold">500 ERR</span>
									<span class="text-slate-500">3.10s</span>
								</div>
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}

export default App;
