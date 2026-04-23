<script lang="ts">
	import {
		Beaker,
		BuildingLibrary,
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
	import cytoscape, { type EdgeDefinition, type NodeDefinition } from 'cytoscape';
	// @ts-expect-error no type declarations for cytoscape-cola
	import cola from 'cytoscape-cola';
	import { onMount } from 'svelte';

	interface Props {
		edges: EdgeDefinition[];
		nodes: NodeDefinition[];
		initialSelect?: string | null;
		overlay?: boolean;
		onselect?: (node: NodeDefinition | null) => void;
	}

	let { edges, nodes, initialSelect = null, overlay = false, onselect }: Props = $props();

	let graphEl: HTMLElement;
	let cy: cytoscape.Core;

	onMount(() => {
		cytoscape.use(cola);

		cy = cytoscape({
			container: graphEl,
			layout: { name: 'cola', infinite: true, fit: false } as never,
			style: buildStyle(overlay) as never,
			elements: { nodes, edges },
			minZoom: 0.5,
			maxZoom: 2,
			wheelSensitivity: 0.6
		});

		if (initialSelect) {
			cy.nodes(`node[id="${initialSelect}"]`).select();
		}

		cy.on('tap', (e) => {
			const target = e.target;
			if (target === cy) {
				onselect?.(null);
			} else {
				onselect?.({ group: 'nodes', data: { id: target.id(), ...target.data() } });
			}
		});
	});

	$effect(() => {
		if (cy) cy.style(buildStyle(overlay) as never);
	});

	$effect(() => {
		if (!cy) return;
		cy.elements().remove();
		cy.add([...nodes, ...edges]);
		cy.layout({ name: 'cola', infinite: true, fit: false } as never).run();
	});

	export function center() {
		cy?.reset();
	}

	function svgIcon(icon: IconSource, color: string): string {
		return (
			'data:image/svg+xml;utf8,' +
			encodeURIComponent(
				`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="${color}">` +
					(icon.default.path ?? [])
						.map(
							(p) =>
								`<path ${Object.entries(p)
									.map(([k, v]) => `${k}="${v}"`)
									.join(' ')} />`
						)
						.join('') +
					`</svg>`
			)
		);
	}

	const CONFIRMATE = '#005B99';

	function nodeStyle(type: string, icon: IconSource, withOverlay: boolean): object[] {
		const base: object[] = [
			{
				selector: `node[type\\.${type}]`,
				style: {
					shape: 'rectangle',
					'background-image': svgIcon(icon, '#111827'),
					'background-fit': 'cover',
					'background-color': 'white'
				}
			},
			{
				selector: `node[type\\.${type}]:selected`,
				style: { shape: 'rectangle', 'background-image': svgIcon(icon, CONFIRMATE), color: CONFIRMATE }
			}
		];

		if (!withOverlay) return base;

		return base.concat([
			{
				selector: `node[type\\.${type}][status=1]`,
				style: {
					shape: 'rectangle',
					'background-image': svgIcon(icon, '#166534'),
					'background-fit': 'cover',
					'background-color': 'white',
					color: '#166534'
				}
			},
			{
				selector: `node[type\\.${type}][status=2]`,
				style: {
					shape: 'rectangle',
					'background-image': svgIcon(icon, '#991b1b'),
					'background-fit': 'cover',
					'background-color': 'white',
					color: '#991b1b'
				}
			},
			{
				selector: `node[type\\.${type}]:selected`,
				style: { shape: 'rectangle', 'background-image': svgIcon(icon, CONFIRMATE), color: CONFIRMATE }
			}
		]);
	}

	function buildStyle(withOverlay: boolean): object[] {
		return [
			{
				selector: 'edge',
				style: {
					width: 1,
					'line-color': '#d1d5db',
					'target-arrow-color': '#d1d5db',
					'target-arrow-shape': 'triangle',
					'arrow-scale': 0.8,
					'curve-style': 'bezier'
				}
			},
			{
				selector: 'node',
				style: {
					content: 'data(label)',
					'font-family': 'ui-sans-serif, system-ui, sans-serif',
					'font-size': '13px',
					'text-background-color': 'white',
					'text-background-shape': 'rectangle',
					'text-background-opacity': 1,
					'text-wrap': 'ellipsis',
					'text-max-width': '100px',
					'text-margin-x': 0,
					'text-margin-y': -2,
					color: '#111827'
				}
			},
			...nodeStyle('Storage', CircleStack, withOverlay),
			...nodeStyle('ResourceGroup', RectangleGroup, withOverlay),
			...nodeStyle('Account', Cloud, withOverlay),
			...nodeStyle('Networking', Share, withOverlay),
			...nodeStyle('NetworkService', ServerStack, withOverlay),
			...nodeStyle('Compute', CpuChip, withOverlay),
			...nodeStyle('VirtualMachine', ComputerDesktop, withOverlay),
			...nodeStyle('Function', CommandLine, withOverlay),
			...nodeStyle('Application', CodeBracketSquare, withOverlay),
			...nodeStyle('Library', BuildingLibrary, withOverlay),
			...nodeStyle('TranslationUnitDeclaration', CodeBracket, withOverlay),
			...nodeStyle('CodeRepository', Folder, withOverlay),
			...nodeStyle('KeyVault', LockClosed, withOverlay),
			...nodeStyle('Key', Key, withOverlay),
			...nodeStyle('Secret', PuzzlePiece, withOverlay),
			...nodeStyle('Certificate', Newspaper, withOverlay),
			...nodeStyle('Object', Document, withOverlay),
			...nodeStyle('NetworkSecurityGroup', ShieldCheck, withOverlay),
			...nodeStyle('MLWorkspace', Beaker, withOverlay)
		];
	}
</script>

<div class="h-[calc(100vh-22rem)] w-full" bind:this={graphEl}></div>
