import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
	plugins: [tailwindcss(), solidPlugin()],
	server: {
		port: 5173,
		proxy: {
			"/run": "http://localhost:8080",
			"/stream": "http://localhost:8080",
		},
	},
	build: {
		outDir: "../cmd/server/static",
		emptyOutDir: true,
	},
});
