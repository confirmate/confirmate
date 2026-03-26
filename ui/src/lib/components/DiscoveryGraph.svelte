<script lang="ts" module>
	import { writable } from 'svelte/store';
	export const shouldCenter = writable<Boolean>(false);
</script>

<script lang="ts">
	import { run } from 'svelte/legacy';

	import {
		Beaker,
		BuildingLibrary,
		ChartPie,
		CircleStack,
		Cloud,
		CodeBracket,
		CodeBracketSquare,
		CommandLine,
		ComputerDesktop,
		CpuChip,
		Document,
		Folder,
		Key,
		LockClosed,
		Newspaper,
		PuzzlePiece,
		RectangleGroup,
		ServerStack,
		Share,
		ShieldCheck
	} from '@steeze-ui/heroicons';
	import type { IconSource } from '@steeze-ui/svelte-icon';
	import cytoscape, { type EdgeDefinition, type NodeDefinition, type Stylesheet } from 'cytoscape';
	import cola from 'cytoscape-cola';
	import { createEventDispatcher, onMount, setContext } from 'svelte';

	interface Props {
		edges: EdgeDefinition[];
		nodes: NodeDefinition[];
		initialSelect: string | null;
		overlay: boolean;
	}

	let { edges, nodes, initialSelect, overlay }: Props = $props();

	let graph: HTMLElement = $state();
	let cy: cytoscape.Core = $state();

	const dispatch = createEventDispatcher<{
		select: NodeDefinition | null;
	}>();

	interface $$Events {
		select: CustomEvent<NodeDefinition | null>;
	}

	setContext('graphSharedState', {
		getCyInstance: () => cy
	});

	onMount(() => {
		cytoscape.use(cola);

		cy = cytoscape({
			container: graph,
			layout: {
				name: 'cola',
				infinite: true,
				fit: false
			},
			style: style(overlay),
			elements: {
				nodes: nodes,
				edges: edges
			},
			minZoom: 0.5,
			maxZoom: 2,
			wheelSensitivity: 0.6
		});

		if (initialSelect) {
			var node = cy.nodes(`node[id="${initialSelect}"]`);
			node.select();
		}

		cy.on('tap', function (e) {
			var target = e.target;

			if (target === cy) {
				dispatch('select', null);
			} else {
				dispatch('select', { id: target.id(), data: target.data() });
			}
		});
	});

	function svg(icon: IconSource, color: string): string {
		/*
    var svg = document.createElement('svg');
    svg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('stroke-width', '1.5');
    svg.setAttribute('stroke', color);
    svg.setAttribute('class', 'w-6 h-6');
    var path = document.createElement('path');
    for (let p of icon.default.path ?? []) {
      Object.entries(p).forEach(([k, v]) => {
        path.setAttribute(k, v);
      });
    }
    svg.append(path);
    console.log(svg.outerHTML);

    let output = svg.outerHTML;
    return 'data:image/svg+xml;utf8,' + encodeURIComponent(output);
    */
		return (
			'data:image/svg+xml;utf8,' +
			encodeURIComponent(
				`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="${color}">
          ${icon.default.path?.map(
						(p) =>
							`<path ${Object.entries(p)
								.map(([k, v]) => `${k}="${v}"`)
								.join(' ')} />`
					)}</svg>`
			)
		);
	}

	function style(overlay: boolean): Stylesheet[] {
		let styles: Stylesheet[] = [];
		styles.push({
			selector: 'node',
			style: {
				content: `data(label)`,
				'font-family': `"Inter var", sans-serif`,
				'font-size': '1em',
				'text-background-color': 'white',
				'text-background-shape': 'rectangle',
				'text-background-opacity': 1,
				'text-wrap': 'ellipsis',
				'text-max-width': '100px',
				'text-margin-x': 0,
				'text-margin-y': -2,
				color: '#111827'
			}
		});

		styles = styles.concat([
			...nodeStyle('Storage', CircleStack, overlay),
			...nodeStyle('ResourceGroup', RectangleGroup, overlay),
			...nodeStyle('Account', Cloud, overlay),
			...nodeStyle('Networking', Share, overlay),
			...nodeStyle('NetworkService', ServerStack, overlay),
			...nodeStyle('Compute', CpuChip, overlay),
			...nodeStyle('VirtualMachine', ComputerDesktop, overlay),
			...nodeStyle('Function', CommandLine, overlay),
			...nodeStyle('Application', CodeBracketSquare, overlay),
			...nodeStyle('Library', BuildingLibrary, overlay),
			...nodeStyle('TranslationUnitDeclaration', CodeBracket, overlay),
			...nodeStyle('CodeRepository', Folder, overlay),
			...nodeStyle('KeyVault', LockClosed, overlay),
			...nodeStyle('Key', Key, overlay),
			...nodeStyle('Secret', PuzzlePiece, overlay),
			...nodeStyle('Certificate', Newspaper, overlay),
			...nodeStyle('Object', Document, overlay),
			...nodeStyle('NetworkSecurityGroup', ShieldCheck, overlay),
			...nodeStyle('MLWorkspace', Beaker, overlay)
		]);

		return styles;
	}

	function nodeStyle(type: string, icon: IconSource, overlay: boolean): Stylesheet[] {
		let styles: Stylesheet[] = [
			{
				selector: `node[type\\.${type}]`,
				style: {
					shape: 'rectangle',
					'background-image': svg(icon, '#111827'),
					'background-fit': 'cover',
					'background-color': 'white'
				}
			},
			{
				selector: `node[type\\.${type}]:selected`,
				style: {
					shape: 'rectangle',
					'background-image': svg(icon, '#007FC3'),
					color: '#007FC3'
				}
			}
		];

		if (overlay) {
			styles = styles.concat([
				{
					selector: `node[type\\.${type}][status=1]`,
					style: {
						shape: 'rectangle',
						'background-image': svg(icon, '#166534'),
						'background-fit': 'cover',
						'background-color': 'white',
						color: '#166534'
					}
				},
				{
					selector: `node[type\\.${type}][status=2]`,
					style: {
						shape: 'rectangle',
						'background-image': svg(icon, '#991b1b'),
						'background-fit': 'cover',
						'background-color': 'white',
						color: '#991b1b'
					}
				},
				{
					selector: `node[type\\.${type}]:selected`,
					style: {
						shape: 'rectangle',
						'background-image': svg(icon, '#007FC3'),
						color: '#007FC3'
					}
				}
			]);
		}

		return styles;
	}

	run(() => {
		(() => {
			if (cy) {
				cy.style(style(overlay));

				if ($shouldCenter) {
					cy.reset();
					$shouldCenter = false;
				}
			}
		})();
	});
</script>

<div class="graph h-[calc(100vh-25rem)] max-w-7xl" bind:this={graph}></div>
