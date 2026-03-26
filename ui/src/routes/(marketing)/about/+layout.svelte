<script lang="ts">
	import { CloudArrowUp, LockOpen, PuzzlePiece } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();

	const features = [
		{
			name: 'Future-proof.',
			path: '/about/page1',
			description:
				"Today's security landscape is changing. The EU Cyber Resilience Act will change how European companies will do business in the cloud. Confirmate is already equipped with rules for the latest EU certifications.",
			icon: CloudArrowUp
		},
		{
			name: 'Avoid vendor lock-in.',
			path: '/about/page2',
			description:
				'Using our ontology approach, we can easily target multi-cloud environments and our rule engine is equipped with mappings from our metrics to different security standards, such as EUCS or BSI C5.',
			icon: LockOpen
		},
		{
			name: 'Holistic.',
			path: '/about/page3',
			description:
				'Using integrations with other tools, such as Codyze, we can provide a holistic view of the Target of Evaluation, from infrastructure to code.',
			icon: PuzzlePiece
		}
	];
</script>

<div class="overflow-hidden bg-white py-24 sm:py-24">
	<div class="mx-auto max-w-7xl px-6 lg:px-8">
		<div
			class="mx-auto grid max-w-2xl grid-cols-1 gap-x-8 gap-y-16 sm:gap-y-20 lg:mx-0 lg:max-w-none lg:grid-cols-2"
		>
			<div class="lg:pt-4 lg:pr-8">
				<div class="lg:max-w-lg">
					<h2 class="text-confirmate text-base leading-7 font-semibold">Be audit-ready, any time</h2>
					<p class="mt-2 text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
						<a href="/about/start">Confirmate</a>
					</p>
					<p class="mt-6 text-lg leading-8 text-gray-600">
						Confirmate, developed by Fraunhofer AISEC, is an open-source cloud compliance tool that
						helps security professionals to get a continuous evaluation of their certification
						target.
					</p>
					<dl class="mt-10 max-w-xl space-y-8 text-base leading-7 text-gray-600 lg:max-w-none">
						{#each features as feature (feature.name)}
							<div class="relative pl-9">
								<dt class="inline font-semibold text-gray-900">
									<Icon
										src={feature.icon}
										class="text-confirmate absolute top-1 left-1 h-5 w-5"
										aria-hidden="true"
									/>
									<a href={feature.path}>{feature.name}</a>
								</dt>
								{' '}
								<dd class="inline">{feature.description}</dd>
							</div>
						{/each}
					</dl>
				</div>
			</div>
			{@render children?.()}
		</div>
	</div>
</div>
