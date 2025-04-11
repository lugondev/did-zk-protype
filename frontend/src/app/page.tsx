'use client'

import DIDForm from '@/components/DIDForm'

export default function Home() {
	return (
		<main className='min-h-screen bg-gray-100 py-10'>
			<div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
				<div className='text-center mb-10'>
					<h1 className='text-4xl font-bold text-gray-900 mb-4'>Decentralized Identity (DID) System</h1>
					<p className='text-lg text-gray-600'>Create and verify your decentralized identity using zero-knowledge proofs</p>
				</div>
				<DIDForm />
			</div>
		</main>
	)
}
