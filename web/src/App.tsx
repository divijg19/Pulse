import { createSignal } from "solid-js";

function App() {
	const [activeTab, setActiveTab] = createSignal("timeline");

	return (
		<div class="h-screen w-screen bg-[#09090b] text-zinc-300 font-sans overflow-hidden flex flex-col relative selection:bg-cyan-500/30">
			{/* Premium Background: Dark vignette over an engineering grid */}
			<div class="absolute inset-0 bg-grid-pattern z-0"></div>
			<div class="absolute inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_0%,transparent_0%,#09090b_100%)] z-0"></div>

			{/* Main Layout Container */}
			<div class="relative z-10 flex flex-col h-full w-full max-w-6xl mx-auto px-6 py-8 gap-8">
				{/* 1. TOP NAV / BRANDING */}
				<header class="flex justify-between items-center shrink-0">
					<div class="flex items-center gap-3">
						{/* Minimal glowing status orb */}
						<div class="relative flex h-2.5 w-2.5">
							<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-60"></span>
							<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.8)]"></span>
						</div>
						<h1 class="text-xl font-bold tracking-tight text-zinc-100">
							Pulse
						</h1>
					</div>
					<div class="text-[10px] uppercase tracking-[0.2em] text-zinc-500 font-mono">
						System Ready
					</div>
				</header>

				{/* 2. THE FOCAL POINT: The Command Bar */}
				{/* We use a perfectly proportioned h-14 bar. Everything is horizontally aligned. */}
				<div class="shrink-0 w-full max-w-3xl mx-auto flex flex-col gap-4">
					<div class="relative group">
						{/* Animated Glow behind the input bar */}
						<div class="absolute -inset-0.5 bg-linear-to-r from-cyan-500/20 via-blue-500/20 to-purple-500/20 rounded-xl blur opacity-30 group-hover:opacity-60 transition duration-500"></div>

						<div class="relative flex items-center bg-[#111113] border border-white/10 rounded-xl h-14 p-1 shadow-2xl focus-within:border-cyan-500/50 transition-colors">
							<select class="h-full bg-transparent border-none text-cyan-400 font-bold px-4 outline-none cursor-pointer appearance-none text-sm tracking-widest uppercase hover:text-cyan-300 transition-colors">
								<option class="bg-zinc-900 text-cyan-400">GET</option>
								<option class="bg-zinc-900 text-green-400">POST</option>
								<option class="bg-zinc-900 text-orange-400">PUT</option>
								<option class="bg-zinc-900 text-rose-400">DELETE</option>
							</select>

							<div class="w-px h-6 bg-white/10 mx-2"></div>

							<input
								type="text"
								value="https://httpbin.org/delay/1"
								class="flex-1 min-w-0 bg-transparent border-none px-2 outline-none font-mono text-zinc-200 placeholder-zinc-600 text-[15px]"
								placeholder="Enter API endpoint..."
							/>

							<div class="w-px h-6 bg-white/10 mx-2"></div>

							<div class="flex items-center gap-2 px-3 hover:bg-white/5 rounded-lg transition-colors h-full cursor-text flex-none w-36 justify-center">
								<span class="text-zinc-500 text-xs font-mono uppercase tracking-widest">
									CC
								</span>
								<input
									type="number"
									value="10"
									class="bg-transparent w-14 outline-none text-zinc-200 font-mono text-[15px] text-center"
								/>
							</div>

							{/* The Action Button: Proportionate, not massive. Inner shadow for depth. */}
							<button
								type="button"
								class="h-full px-8 ml-2 rounded-lg bg-zinc-100 text-zinc-950 font-bold text-sm tracking-widest uppercase hover:bg-white transition-all shadow-[inset_0_-2px_4px_rgba(0,0,0,0.2)] active:shadow-[inset_0_2px_4px_rgba(0,0,0,0.2)] active:translate-y-px"
							>
								Run
							</button>
						</div>
					</div>

					{/* 3. AMBIENT METRICS STRIP */}
					<div class="flex justify-between items-center px-4 py-2 bg-linear-to-r from-transparent via-white/2 to-transparent border-y border-white/2">
						{[
							{ label: "Requests", val: "100" },
							{ label: "Success", val: "100%", textClass: "text-emerald-400" },
							{
								label: "Avg Latency",
								val: "1.24s",
								textClass: "text-amber-400",
							},
							{ label: "RPS", val: "8.5" },
						].map((m) => (
							<div class="flex items-baseline gap-2">
								<span class="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
									{m.label}
								</span>
								<span
									class={`font-mono text-sm ${m.textClass || "text-zinc-300"}`}
								>
									{m.val}
								</span>
							</div>
						))}
					</div>
				</div>

				{/* 4. THE DATA WORKSPACE */}
				<div class="flex-1 min-h-0 flex flex-col bg-[#0f0f11]/80 backdrop-blur-xl border border-white/5 rounded-2xl overflow-hidden shadow-2xl relative">
					{/* Workspace Header / Segmented Control */}
					<div class="flex items-center justify-center p-3 border-b border-white/5 bg-white/1">
						<div class="flex p-1 bg-black/40 rounded-lg border border-white/5 backdrop-blur-md">
							<button
								type="button"
								onClick={() => setActiveTab("timeline")}
								class={`px-6 py-1.5 rounded-md text-xs font-bold uppercase tracking-widest transition-all duration-300 ${activeTab() === "timeline" ? "bg-zinc-800 text-white shadow-md" : "text-zinc-500 hover:text-zinc-300"}`}
							>
								Timeline
							</button>
							<button
								type="button"
								onClick={() => setActiveTab("logs")}
								class={`px-6 py-1.5 rounded-md text-xs font-bold uppercase tracking-widest transition-all duration-300 ${activeTab() === "logs" ? "bg-zinc-800 text-white shadow-md" : "text-zinc-500 hover:text-zinc-300"}`}
							>
								Live Logs
							</button>
						</div>
					</div>

					{/* Workspace Content */}
					<div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
						{/* TIMELINE VIEW */}
						{activeTab() === "timeline" && (
							<div class="space-y-1.5 max-w-4xl mx-auto w-full">
								{/* Dummy Bar 1 */}
								<div class="w-full bg-black/40 rounded h-8 relative overflow-hidden flex items-center px-3 group border border-transparent hover:border-white/5 transition-colors">
									{/* The actual latency bar */}
									<div class="absolute left-0 top-0 bottom-0 bg-cyan-500/10 w-[60%] border-r border-cyan-400 group-hover:bg-cyan-500/20 transition-colors"></div>

									<div class="relative z-10 flex justify-between w-full text-xs font-mono">
										<span class="text-zinc-400 group-hover:text-zinc-200">
											200 OK
										</span>
										<span class="text-cyan-400">1.24s</span>
									</div>
								</div>
								{/* Dummy Bar 2 */}
								<div class="w-full bg-black/40 rounded h-8 relative overflow-hidden flex items-center px-3 group border border-transparent hover:border-white/5 transition-colors">
									<div class="absolute left-0 top-0 bottom-0 bg-rose-500/10 w-[85%] border-r border-rose-500 group-hover:bg-rose-500/20 transition-colors"></div>
									<div class="relative z-10 flex justify-between w-full text-xs font-mono">
										<span class="text-zinc-400 group-hover:text-zinc-200">
											500 ERR
										</span>
										<span class="text-rose-400">3.10s</span>
									</div>
								</div>
							</div>
						)}

						{/* LOGS VIEW */}
						{activeTab() === "logs" && (
							<div class="font-mono text-xs space-y-1 max-w-4xl mx-auto w-full">
								<div class="flex gap-4 p-1.5 rounded hover:bg-white/5 transition-colors items-center group">
									<span class="text-zinc-600 w-24">12:01:00.145</span>
									<span class="text-cyan-400 w-10">GET</span>
									<span class="text-zinc-400 truncate flex-1 group-hover:text-zinc-200 transition-colors">
										https://httpbin.org/delay/1
									</span>
									<span class="text-cyan-400 w-16 text-right">200</span>
									<span class="text-zinc-500 w-16 text-right">1.24s</span>
								</div>
								<div class="flex gap-4 p-1.5 rounded bg-rose-500/5 border border-rose-500/10 hover:bg-rose-500/10 transition-colors items-center group">
									<span class="text-zinc-600 w-24">12:01:01.320</span>
									<span class="text-rose-400 w-10">GET</span>
									<span class="text-zinc-400 truncate flex-1 group-hover:text-zinc-200 transition-colors">
										https://httpbin.org/delay/1
									</span>
									<span class="text-rose-500 w-16 text-right font-bold">
										500
									</span>
									<span class="text-rose-400/70 w-16 text-right">3.10s</span>
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
