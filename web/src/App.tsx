import { createEffect, createSignal, For, onCleanup } from "solid-js";

// Define what the Go backend sends us
type Result = {
	Status: number;
	Latency: number; // Go sends time.Duration as nanoseconds
	Error: string;
};

export default function App() {
	const [activeTab, setActiveTab] = createSignal("timeline");

	// Input State
	const [method, setMethod] = createSignal("GET");
	const [url, setUrl] = createSignal("https://httpbin.org/delay/1");
	const [concurrency, setConcurrency] = createSignal(10);

	// Execution State
	const [isRunning, setIsRunning] = createSignal(false);
	const [results, setResults] = createSignal<Result[]>([]);

	// Derived State (Max Latency for the Timeline Bars)
	const maxLatency = () => {
		const res = results();
		if (res.length === 0) return 0;
		return Math.max(...res.map((r) => r.Latency));
	};

	// 🔌 Establish the SSE Stream Connection
	createEffect(() => {
		// The Vite proxy routes this to http://localhost:8080/stream
		const eventSource = new EventSource("/stream");

		eventSource.addEventListener("result", (event) => {
			const data = JSON.parse(event.data) as Result;
			// Append the new result to our state array
			setResults((prev) => [...prev, data]);
		});

		eventSource.onerror = (err) => {
			console.error("SSE Error:", err);
		};

		// Cleanup when component unmounts
		onCleanup(() => eventSource.close());
	});

	// 🚀 Fire the Request Batch
	const handleRun = async () => {
		if (isRunning()) return;

		setIsRunning(true);
		setResults([]); // Clear previous run

		try {
			await fetch("/run", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					url: url(),
					method: method(),
					concurrency: Number(concurrency()),
				}),
			});
			// We don't need to wait for the response to render data!
			// The EventSource will catch the stream and populate the UI automatically.
		} catch (e) {
			console.error("Failed to start run", e);
		} finally {
			// Small timeout to let the final SSE events arrive before resetting button
			setTimeout(() => setIsRunning(false), 500);
		}
	};

	// Helper to format nanoseconds to seconds (e.g., 1.24s)
	const formatLatency = (nanoseconds: number) => {
		return `${(nanoseconds / 1_000_000_000).toFixed(2)}s`;
	};

	return (
		<div class="h-screen w-screen bg-[#09090b] text-zinc-300 font-sans overflow-hidden flex flex-col relative selection:bg-cyan-500/30">
			<div class="absolute inset-0 bg-grid-pattern z-0"></div>
			<div class="absolute inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_0%,transparent_0%,#09090b_100%)] z-0"></div>

			<div class="relative z-10 flex flex-col h-full w-full max-w-6xl mx-auto px-6 py-8 gap-8">
				{/* HEADER */}
				<header class="flex justify-between items-center shrink-0">
					<div class="flex items-center gap-3">
						<div class="relative flex h-2.5 w-2.5">
							<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-60"></span>
							<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-cyan-500 shadow-[0_0_8px_rgba(6,182,212,0.8)]"></span>
						</div>
						<h1 class="text-xl font-bold tracking-tight text-zinc-100">
							Pulse
						</h1>
					</div>
					<div class="text-[10px] uppercase tracking-[0.2em] text-zinc-500 font-mono">
						{isRunning() ? (
							<span class="text-cyan-400 animate-pulse">Running Batch...</span>
						) : (
							"System Ready"
						)}
					</div>
				</header>

				{/* COMMAND BAR */}
				<div class="shrink-0 w-full max-w-3xl mx-auto flex flex-col gap-4">
					<div class="relative group">
						<div class="absolute -inset-0.5 bg-linear-to-r from-cyan-500/20 via-blue-500/20 to-purple-500/20 rounded-xl blur opacity-30 group-hover:opacity-60 transition duration-500"></div>

						<div class="relative flex items-center bg-[#111113] border border-white/10 rounded-xl h-14 p-1 shadow-2xl focus-within:border-cyan-500/50 transition-colors">
							<select
								value={method()}
								onChange={(e) => setMethod(e.currentTarget.value)}
								class="h-full bg-transparent border-none text-cyan-400 font-bold px-4 outline-none cursor-pointer appearance-none text-sm tracking-widest uppercase hover:text-cyan-300 transition-colors"
							>
								<option class="bg-zinc-900 text-cyan-400">GET</option>
								<option class="bg-zinc-900 text-green-400">POST</option>
								<option class="bg-zinc-900 text-orange-400">PUT</option>
								<option class="bg-zinc-900 text-rose-400">DELETE</option>
							</select>

							<div class="w-px h-6 bg-white/10 mx-1"></div>

							<input
								type="text"
								value={url()}
								onInput={(e) => setUrl(e.currentTarget.value)}
								class="flex-1 min-w-0 bg-transparent border-none px-2 outline-none font-mono text-zinc-200 placeholder-zinc-600 text-[15px]"
								placeholder="Enter API endpoint..."
							/>

							<div class="w-px h-6 bg-white/10 mx-1"></div>

							<div class="flex items-center gap-3 px-3 hover:bg-white/5 rounded-lg transition-colors h-full cursor-text flex-none w-44 justify-center">
								<span class="text-zinc-300 text-sm font-mono uppercase tracking-widest">
									CC
								</span>
								<input
									type="number"
									value={concurrency()}
									onInput={(e) => setConcurrency(Number(e.currentTarget.value))}
									class="bg-transparent w-16 outline-none text-zinc-200 font-mono text-[15px] text-center"
								/>
							</div>

							<button
								type="button"
								onClick={handleRun}
								disabled={isRunning()}
								class={`h-full px-6 ml-1 rounded-lg font-bold text-sm tracking-widest uppercase transition-all shadow-[inset_0_-2px_4px_rgba(0,0,0,0.2)] ${
									isRunning()
										? "bg-zinc-800 text-zinc-600 cursor-not-allowed"
										: "bg-zinc-100 text-zinc-950 hover:bg-white active:translate-y-px active:shadow-[inset_0_2px_4px_rgba(0,0,0,0.2)]"
								}`}
							>
								{isRunning() ? "..." : "Run"}
							</button>
						</div>
					</div>
				</div>

				{/* WORKSPACE */}
				<div class="flex-1 min-h-0 flex flex-col bg-[#0f0f11]/80 backdrop-blur-xl border border-white/5 rounded-2xl overflow-hidden shadow-2xl relative">
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

					<div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
						{activeTab() === "timeline" && (
							<div class="space-y-1.5 max-w-4xl mx-auto w-full">
								{results().length === 0 && (
									<div class="text-center text-zinc-600 font-mono mt-10">
										Awaiting execution...
									</div>
								)}

								{/* 📊 DYNAMIC TIMELINE BARS */}
								<For each={results()}>
									{(res) => {
										const isError = res.Status >= 400 || res.Status === 0;
										const width = Math.max(
											(res.Latency / maxLatency()) * 100,
											1,
										); // Min 1% width

										return (
											<div class="w-full bg-black/40 rounded h-8 relative overflow-hidden flex items-center px-3 group border border-transparent hover:border-white/5 transition-colors">
												<div
													class={`absolute left-0 top-0 bottom-0 transition-all duration-300 ${isError ? "bg-rose-500/10 border-r border-rose-500 group-hover:bg-rose-500/20" : "bg-cyan-500/10 border-r border-cyan-400 group-hover:bg-cyan-500/20"}`}
													style={{ width: `${width}%` }}
												></div>
												<div class="relative z-10 flex justify-between w-full text-xs font-mono">
													<span class="text-zinc-400 group-hover:text-zinc-200">
														{res.Status === 0 ? "FAILED" : `${res.Status} OK`}
													</span>
													<span
														class={isError ? "text-rose-400" : "text-cyan-400"}
													>
														{formatLatency(res.Latency)}
													</span>
												</div>
											</div>
										);
									}}
								</For>
							</div>
						)}

						{activeTab() === "logs" && (
							<div class="font-mono text-xs space-y-1 max-w-4xl mx-auto w-full">
								{results().length === 0 && (
									<div class="text-center text-zinc-600 mt-10">
										No logs yet.
									</div>
								)}

								{/* 📜 DYNAMIC LOGS */}
								<For each={results()}>
									{(res) => {
										const isError = res.Status >= 400 || res.Status === 0;
										return (
											<div
												class={`flex gap-4 p-1.5 rounded items-center group transition-colors ${isError ? "bg-rose-500/5 border border-rose-500/10 hover:bg-rose-500/10" : "hover:bg-white/5"}`}
											>
												<span
													class={
														isError
															? "text-rose-400 w-10"
															: "text-cyan-400 w-10"
													}
												>
													{method()}
												</span>
												<span class="text-zinc-400 truncate flex-1 group-hover:text-zinc-200">
													{url()}
												</span>
												<span
													class={`w-16 text-right font-bold ${isError ? "text-rose-500" : "text-cyan-400"}`}
												>
													{res.Status === 0 ? "ERR" : res.Status}
												</span>
												<span
													class={`w-16 text-right ${isError ? "text-rose-400/70" : "text-zinc-500"}`}
												>
													{formatLatency(res.Latency)}
												</span>
											</div>
										);
									}}
								</For>
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}
