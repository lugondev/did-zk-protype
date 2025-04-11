import type { Config } from 'tailwindcss'

import plugin from '@tailwindcss/forms'

const config: Config = {
	content: [
		'./src/pages/**/*.{js,ts,jsx,tsx,mdx}',
		'./src/components/**/*.{js,ts,jsx,tsx,mdx}',
		'./src/app/**/*.{js,ts,jsx,tsx,mdx}',
	],
	theme: {
		extend: {
			colors: {
				gray: {
					50: '#f9fafb',
					100: '#f3f4f6',
					200: '#e5e7eb',
					300: '#d1d5db',
					400: '#9ca3af',
					500: '#6b7280',
					600: '#4b5563',
					700: '#374151',
					800: '#1f2937',
					900: '#111827',
				},
				indigo: {
					500: '#6366f1',
					600: '#4f46e5',
					700: '#4338ca',
				},
				green: {
					500: '#22c55e',
					600: '#16a34a',
					700: '#15803d',
				},
				red: {
					50: '#fef2f2',
					200: '#fecaca',
					600: '#dc2626',
				},
			},
		},
	},
	plugins: [plugin],
}

export default config
