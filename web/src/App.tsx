import { createEffect, createSignal, For, onCleanup, Show } from "solid-js";

type Result = {
	Status: number;
	Latency: number;
	Error: string;
	ResponseHeaders?: Record<string, string>;
	ResponseBody?: string;
	RequestMethod?: string;
	RequestURL?: string;
};

export default function App() {
	const [activeTab, setActiveTab] = createSignal("timeline");

	const [method, setMethod] = createSignal("GET");
	const [url, setUrl] = createSignal("https://httpbin.org/delay/1");
	const [concurrency, setConcurrency] = createSignal(10);

	const [isRunning, setIsRunning] = createSignal(false);
	const [results, setResults] = createSignal<Result[]>([]);
	const [reqHeaders, setReqHeaders] = createSignal<
		{ key: string; value: string }[]
	>([]);
	const [reqBody, setReqBody] = createSignal("");
	const [showPayloadEditor, setShowPayloadEditor] = createSignal(false);
	const [selectedResult, setSelectedResult] = createSignal<Result | null>(null);
	const [activeRequestMethod, setActiveRequestMethod] = createSignal("GET");

	// LIVE TIMER STATE
	const [elapsedMs, setElapsedMs] = createSignal(0);

	// --- 🧮 LIVE MATH & METRICS ---
	const maxLatency = () => {
		const res = results();
		if (res.length === 0) return 0;
		return Math.max(...res.map((r) => r.Latency));
	};

	const totalReqs = () => results().length;

	const successRate = () => {
		const total = totalReqs();
		if (total === 0) return "0%";
		const successes = results().filter(
			(r) => r.Status >= 200 && r.Status < 400,
		).length;
		return `${Math.round((successes / total) * 100)}%`;
	};

	const avgLatency = () => {
		const total = totalReqs();
		if (total === 0) return "0.00s";
		const sum = results().reduce((acc, r) => acc + r.Latency, 0);
		return formatLatency(sum / total);
	};

	const currentRPS = () => {
		const total = totalReqs();
		const seconds = elapsedMs() / 1000;
		if (seconds === 0 || total === 0) return "0.0";
		return (total / seconds).toFixed(1);
	};

	// --- 🔌 CONNECTION ---
	createEffect(() => {
		const eventSource = new EventSource("/stream");
		eventSource.addEventListener("result", (event) => {
			const data = JSON.parse(event.data) as Result;
			setResults((prev) => [
				...prev,
				{
					...data,
					RequestMethod: activeRequestMethod(),
					RequestURL: url(),
				},
			]);
		});
		eventSource.onerror = (err) => console.error("SSE Error:", err);
		onCleanup(() => eventSource.close());
	});

	const addHeader = () => {
		setReqHeaders((prev) => [...prev, { key: "", value: "" }]);
	};

	const updateHeader = (
		index: number,
		field: "key" | "value",
		value: string,
	) => {
		setReqHeaders((prev) =>
			prev.map((header, i) =>
				i === index ? { ...header, [field]: value } : header,
			),
		);
	};

	const removeHeader = (index: number) => {
		setReqHeaders((prev) => prev.filter((_, i) => i !== index));
	};

	// --- 🚀 EXECUTION ---
	const handleRun = async () => {
		if (isRunning()) return;

		const parsedHeaders = reqHeaders().reduce<Record<string, string>>(
			(acc, header) => {
				const key = header.key.trim();
				if (key !== "") {
					acc[key] = header.value;
				}
				return acc;
			},
			{},
		);

		setActiveRequestMethod(method());
		setUrl(url());

		setIsRunning(true);
		setResults([]);
		setElapsedMs(0);
		setSelectedResult(null);

		// Start a high-speed live timer (ticks every 50ms)
		const startTime = Date.now();
		const timer = setInterval(() => {
			setElapsedMs(Date.now() - startTime);
		}, 50);

		try {
			await fetch("/run", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					url: url(),
					method: method(),
					headers: parsedHeaders,
					body: reqBody(),
					concurrency: Number(concurrency()),
				}),
			});
		} catch (e) {
			console.error("Failed to start run", e);
		} finally {
			setTimeout(() => {
				setIsRunning(false);
				clearInterval(timer); // Stop the clock when done!
			}, 500);
		}
	};

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
							<span
								class={`absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-60 ${isRunning() ? "animate-ping" : ""}`}
							></span>
							<span
								class={`relative inline-flex rounded-full h-2.5 w-2.5 bg-cyan-500 ${isRunning() ? "shadow-[0_0_15px_rgba(6,182,212,1)]" : "shadow-[0_0_8px_rgba(6,182,212,0.5)]"}`}
							></span>
						</div>
						<h1 class="text-xl font-bold tracking-tight text-zinc-100">
							Pulse
						</h1>
					</div>
					<div class="text-[10px] uppercase tracking-[0.2em] font-mono">
						{isRunning() ? (
							<span class="text-cyan-400 animate-pulse">
								{elapsedMs() > 0
									? `${(elapsedMs() / 1000).toFixed(1)}s ELAPSED`
									: "RUNNING..."}
							</span>
						) : (
							<span class="text-zinc-500">SYSTEM READY</span>
						)}
					</div>
				</header>

				{/* COMMAND BAR */}
				<div class="shrink-0 w-full max-w-3xl mx-auto flex flex-col gap-4">
					<div class="relative group">
						<div
							class={`absolute -inset-0.5 rounded-xl blur opacity-30 transition duration-500 ${isRunning() ? "bg-linear-to-r from-cyan-500 via-blue-500 to-cyan-500 opacity-100 animate-pulse" : "bg-linear-to-r from-cyan-500/20 via-blue-500/20 to-purple-500/20 group-hover:opacity-60"}`}
						></div>

						<div class="relative flex items-center bg-[#111113] border border-white/10 rounded-xl h-14 p-2 shadow-2xl focus-within:border-cyan-500/50 transition-colors">
							<select
								value={method()}
								onChange={(e) => setMethod(e.currentTarget.value)}
								class="h-full bg-transparent border-none text-cyan-400 font-bold px-4 outline-none cursor-pointer appearance-none text-sm tracking-widest uppercase hover:text-cyan-300"
							>
								<option class="bg-zinc-900 text-cyan-400">GET</option>
								<option class="bg-zinc-900 text-green-400">POST</option>
								<option class="bg-zinc-900 text-orange-400">PUT</option>
								<option class="bg-zinc-900 text-rose-400">DELETE</option>
							</select>

							<div class="w-px h-6 bg-white/10 mx-3"></div>
							<input
								type="text"
								value={url()}
								onInput={(e) => setUrl(e.currentTarget.value)}
								class="flex-1 min-w-0 bg-transparent border-none px-2 outline-none font-mono text-zinc-200 placeholder-zinc-600 text-[15px]"
								placeholder="Enter API endpoint..."
							/>
							<div class="w-px h-6 bg-white/10 mx-3"></div>

							<div class="flex items-center gap-2 px-2 hover:bg-white/5 transition-colors h-full shrink-0">
								<button
									type="button"
									onClick={() => setConcurrency((c) => Math.max(1, c - 1))}
									class="px-2 py-1 rounded-md text-sm text-zinc-300 hover:bg-white/5 transition-colors cursor-pointer"
								>
									-
								</button>

								<div class="flex items-center gap-2 px-1">
									<span class="text-zinc-500 text-xs font-mono uppercase tracking-widest">
										CC
									</span>
									<input
										type="number"
										value={concurrency()}
										onInput={(e) =>
											setConcurrency(Number(e.currentTarget.value))
										}
										min="1"
										class="no-spinner bg-transparent w-12 outline-none text-zinc-200 font-mono text-[15px] text-center"
									/>
								</div>

								<button
									type="button"
									onClick={() => setConcurrency((c) => c + 1)}
									class="px-2 py-1 rounded-md text-sm text-zinc-300 hover:bg-white/5 transition-colors cursor-pointer"
								>
									+
								</button>

								<button
									type="button"
									onClick={() => setShowPayloadEditor((current) => !current)}
									class={`px-2 py-1 rounded-md text-xs font-mono uppercase tracking-widest border transition-colors cursor-pointer ${showPayloadEditor() ? "text-cyan-300 border-cyan-500/40 bg-cyan-500/10" : "text-zinc-400 border-white/10 hover:text-zinc-200 hover:border-white/20"}`}
								>
									Payload
								</button>
							</div>

							<button
								type="button"
								onClick={handleRun}
								disabled={isRunning()}
								class={`h-full w-28 ml-3 rounded-lg font-bold text-sm tracking-widest uppercase transition-all shadow-[inset_0_-2px_4px_rgba(0,0,0,0.2)] ${isRunning() ? "bg-cyan-900/50 text-cyan-500 cursor-not-allowed" : "bg-zinc-100 text-zinc-950 hover:bg-white active:translate-y-px active:shadow-[inset_0_2px_4px_rgba(0,0,0,0.2)]"}`}
							>
								{isRunning() ? "..." : "Run"}
							</button>
						</div>
					</div>

					<Show when={showPayloadEditor()}>
						<div class="bg-black/20 border-y border-white/5 p-4 rounded-xl">
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								<div class="space-y-3">
									<div class="flex items-center justify-between">
										<h3 class="text-xs font-bold uppercase tracking-widest text-zinc-400">
											Headers
										</h3>
										<button
											type="button"
											onClick={addHeader}
											class="px-2 py-1 rounded-md text-xs font-mono uppercase tracking-widest border border-cyan-500/30 text-cyan-300 hover:bg-cyan-500/10 transition-colors"
										>
											Add Header
										</button>
									</div>

									<div class="space-y-2 max-h-44 overflow-y-auto pr-1">
										<For each={reqHeaders()}>
											{(header, index) => (
												<div class="flex items-center gap-2">
													<input
														type="text"
														value={header.key}
														onInput={(e) =>
															updateHeader(
																index(),
																"key",
																e.currentTarget.value,
															)
														}
														placeholder="Key"
														class="flex-1 h-9 px-2 rounded-md bg-black/40 border border-white/10 text-zinc-200 text-xs font-mono outline-none focus:border-cyan-500/50"
													/>
													<input
														type="text"
														value={header.value}
														onInput={(e) =>
															updateHeader(
																index(),
																"value",
																e.currentTarget.value,
															)
														}
														placeholder="Value"
														class="flex-1 h-9 px-2 rounded-md bg-black/40 border border-white/10 text-zinc-200 text-xs font-mono outline-none focus:border-cyan-500/50"
													/>
													<button
														type="button"
														onClick={() => removeHeader(index())}
														class="h-9 w-9 rounded-md border border-rose-500/30 text-rose-300 hover:bg-rose-500/10 transition-colors"
													>
														x
													</button>
												</div>
											)}
										</For>
									</div>
								</div>

								<div class="space-y-3">
									<h3 class="text-xs font-bold uppercase tracking-widest text-zinc-400">
										Body
									</h3>
									<textarea
										value={reqBody()}
										onInput={(e) => setReqBody(e.currentTarget.value)}
										placeholder='{"name":"pulse"}'
										class="min-h-40 w-full bg-black/40 border border-white/10 text-zinc-300 font-mono text-sm p-2 rounded-lg outline-none focus:border-cyan-500/50"
									/>
								</div>
							</div>
						</div>
					</Show>

					{/* 📊 LIVE METRICS STRIP */}
					<div class="flex justify-between items-center px-4 py-2 bg-linear-to-r from-transparent via-white/2 to-transparent border-y border-white/2">
						<div class="flex items-baseline gap-2 w-1/4">
							<span class="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
								Requests
							</span>
							<span class="font-mono text-sm text-zinc-300">
								{totalReqs()} / {concurrency()}
							</span>
						</div>
						<div class="flex items-baseline gap-2 w-1/4 justify-center">
							<span class="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
								Success
							</span>
							<span
								class={`font-mono text-sm ${parseFloat(successRate()) < 100 ? "text-amber-400" : "text-emerald-400"}`}
							>
								{successRate()}
							</span>
						</div>
						<div class="flex items-baseline gap-2 w-1/4 justify-center">
							<span class="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
								Avg Latency
							</span>
							<span class="font-mono text-sm text-amber-400">
								{avgLatency()}
							</span>
						</div>
						<div class="flex items-baseline gap-2 w-1/4 justify-end">
							<span class="text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
								RPS
							</span>
							<span class="font-mono text-sm text-fuchsia-400">
								{currentRPS()}
							</span>
						</div>
					</div>
				</div>

				{/* WORKSPACE */}
				<div class="flex-1 min-h-0 flex flex-col bg-[#0f0f11]/80 backdrop-blur-xl border border-white/5 rounded-2xl overflow-hidden shadow-2xl relative">
					<div class="flex items-center justify-center p-3 border-b border-white/5 bg-white/1">
						<div class="p-1 bg-black/40 rounded-lg border border-white/5 backdrop-blur-md">
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
							<div class="space-y-1.5 max-w-4xl mx-auto w-full pb-8">
								{results().length === 0 && !isRunning() && (
									<div class="text-center text-zinc-600 font-mono mt-10">
										Awaiting execution...
									</div>
								)}

								<For each={results()}>
									{(res) => {
										const isError = res.Status >= 400 || res.Status === 0;
										const width = Math.max(
											(res.Latency / maxLatency()) * 100,
											1,
										);
										return (
											<button
												type="button"
												onClick={() => setSelectedResult(res)}
												class="w-full bg-black/40 rounded h-8 relative overflow-hidden flex items-center px-3 group border border-transparent hover:border-white/5 transition-colors cursor-pointer"
											>
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
											</button>
										);
									}}
								</For>
							</div>
						)}

						{activeTab() === "logs" && (
							<div class="font-mono text-xs space-y-1 max-w-4xl mx-auto w-full pb-8">
								{results().length === 0 && !isRunning() && (
									<div class="text-center text-zinc-600 mt-10">
										No logs yet.
									</div>
								)}
								<For each={results()}>
									{(res) => {
										const isError = res.Status >= 400 || res.Status === 0;
										return (
											<button
												type="button"
												onClick={() => setSelectedResult(res)}
												class={`flex gap-4 p-1.5 rounded items-center group transition-colors cursor-pointer ${isError ? "bg-rose-500/5 border border-rose-500/10 hover:bg-rose-500/10" : "hover:bg-white/5"}`}
											>
												<span
													class={
														isError
															? "text-rose-400 w-10"
															: "text-cyan-400 w-10"
													}
												>
													{res.RequestMethod || method()}
												</span>
												<span class="text-zinc-400 truncate flex-1 group-hover:text-zinc-200">
													{res.RequestURL || url()}
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
											</button>
										);
									}}
								</For>
							</div>
						)}
					</div>
				</div>

				<div
					class={`fixed top-0 right-0 h-full w-150 max-w-full bg-[#09090b]/95 backdrop-blur-2xl border-l border-white/10 shadow-2xl z-50 transform transition-transform duration-300 ${selectedResult() ? "translate-x-0" : "translate-x-full"}`}
				>
					<div class="h-full flex flex-col">
						<div class="p-4 border-b border-white/10 bg-black/20">
							<div class="flex items-start justify-between gap-4">
								<div class="space-y-2 min-w-0">
									<div class="flex items-center gap-2">
										<span class="text-[10px] uppercase tracking-widest text-zinc-500 font-bold">
											Request
										</span>
										<span class="text-xs font-mono text-cyan-300 uppercase">
											{selectedResult()?.RequestMethod || method()}
										</span>
									</div>
									<div class="text-xs text-zinc-400 font-mono truncate max-w-105">
										{selectedResult()?.RequestURL || url()}
									</div>
									<div class="flex items-center gap-3 text-xs font-mono">
										<span
											class={`${(selectedResult()?.Status || 0) >= 400 || (selectedResult()?.Status || 0) === 0 ? "text-rose-400" : "text-cyan-400"}`}
										>
											Status: {selectedResult()?.Status || 0}
										</span>
										<span class="text-zinc-500">
											Latency: {formatLatency(selectedResult()?.Latency || 0)}
										</span>
									</div>
								</div>
								<button
									type="button"
									onClick={() => setSelectedResult(null)}
									class="px-3 py-2 rounded-md border border-white/15 text-zinc-300 text-xs font-mono uppercase tracking-widest hover:bg-white/5 transition-colors"
								>
									Close
								</button>
							</div>
						</div>

						<div class="flex-1 overflow-y-auto p-4 space-y-4">
							<div class="space-y-2">
								<h3 class="text-[10px] uppercase tracking-widest text-zinc-500 font-bold">
									Response Headers
								</h3>
								<div class="rounded-xl border border-white/10 bg-black/30 divide-y divide-white/5">
									<Show
										when={
											Object.keys(selectedResult()?.ResponseHeaders || {})
												.length > 0
										}
										fallback={
											<div class="p-3 text-xs text-zinc-500 font-mono">
												No headers captured.
											</div>
										}
									>
										<For
											each={Object.entries(
												selectedResult()?.ResponseHeaders || {},
											)}
										>
											{([key, value]) => (
												<div class="p-3 flex flex-col gap-1">
													<span class="text-xs font-mono text-cyan-300 break-all">
														{key}
													</span>
													<span class="text-xs font-mono text-zinc-400 break-all">
														{value}
													</span>
												</div>
											)}
										</For>
									</Show>
								</div>
							</div>

							<Show when={selectedResult()?.Error}>
								<div class="rounded-xl border border-rose-500/30 bg-rose-500/10 p-3 text-xs font-mono text-rose-300 break-all">
									{selectedResult()?.Error}
								</div>
							</Show>

							<div class="space-y-2">
								<h3 class="text-[10px] uppercase tracking-widest text-zinc-500 font-bold">
									Response Body
								</h3>
								<pre class="bg-black/50 p-4 rounded-xl border border-white/10 overflow-x-auto text-xs text-zinc-300 whitespace-pre-wrap wrap-break-word">
									{selectedResult()?.ResponseBody || ""}
								</pre>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
