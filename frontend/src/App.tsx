import {useState, useEffect, useRef, useMemo} from 'react';
import './App.css';
import {GetConfig, SetConfig, InitPaths, FixPaths} from '../wailsjs/go/main/App';
import * as runtime from '../wailsjs/runtime';
import {lib} from '../wailsjs/go/models';
import {useWailsFileDrop} from './lib/wails-drop-interface';

function App() {
	const [messages, setMessages] = useState<(lib.Message | 'divider')[]>([]);
	const [config, setConfig] = useState<lib.AppConfig>({
		onTop: false,
	});
	const receivedCount = useRef(0);
	const hasMessages = messages.length > 0;
	const initDropzoneRef = useRef<HTMLDivElement>(null);
	const fixDropzoneRef = useRef<HTMLDivElement>(null);
	const [isDraggedOver, setIsDraggedOver] = useState(false);
	const hideDropzonesTimeout = useRef(0);

	useEffect(() => {
		const disposers: (() => void)[] = [];
		disposers.push(
			runtime.EventsOn('message', (data: any) =>
				setMessages((messages) => {
					receivedCount.current++;
					return [lib.Message.createFrom(data), ...messages];
				})
			)
		);
		GetConfig().then((config) => setConfig(config));
		disposers.push(runtime.EventsOn('config', (data: any) => setConfig(lib.AppConfig.createFrom(data))));

		return () => {
			disposers.forEach((d) => d());
		};
	}, []);

	function addDivider() {
		setMessages((messages) =>
			messages.length > 0 && messages[0] !== 'divider' ? ['divider', ...messages] : messages
		);
	}

	useWailsFileDrop(initDropzoneRef, (paths) => {
		console.log('init', paths);
		addDivider();
		setIsDraggedOver(false);
		InitPaths(paths).then(console.log, console.error);
	});

	useWailsFileDrop(fixDropzoneRef, (paths) => {
		console.log('fix', paths);
		addDivider();
		setIsDraggedOver(false);
		FixPaths(paths).then(console.log, console.error);
	});

	function handleDragOver() {
		clearTimeout(hideDropzonesTimeout.current);
		setIsDraggedOver(true);
		hideDropzonesTimeout.current = setTimeout(() => setIsDraggedOver(false), 200);
	}

	function logEvent(event: any) {
		if (event.target === event.currentTarget) console.log(event.type);
	}

	return (
		<main className="App" onDragOver={handleDragOver}>
			<section className="messages">
				{messages.map((message, i) =>
					message == 'divider' ? (
						<hr key={receivedCount.current + i} />
					) : (
						<Message key={receivedCount.current + i} data={message} />
					)
				)}
			</section>

			<section
				className={`dropzones ${hasMessages ? '-force-top' : ''} ${
					hasMessages && !isDraggedOver ? '-hide' : ''
				}`}
			>
				<div
					className="dropzone init"
					ref={initDropzoneRef}
					style={{'--wails-drop-target': 'drop'} as React.CSSProperties}
				>
					<h3>Initialize Manager</h3>
					<p>
						Drop <code>.vaj</code> files (or directories containing them) to initialize Clothing Plugins
						Manager inside them.
					</p>
					<p>Other files are ignored. Dropping again only ensures the manager is initialized properly.</p>
				</div>

				<div
					className="dropzone fix"
					ref={fixDropzoneRef}
					style={{'--wails-drop-target': 'drop'} as React.CSSProperties}
				>
					<h3>Fix Files for Release</h3>
					<p>
						Drop the whole package directory inside <code>AddonPackagesBuilder/</code>, and the util will
						fix up everything inside it.
					</p>
					<p>
						Ensures the manager in <code>.vaj</code> files is initialized properly (doesn't initialize files
						that are not), and namespaces clothing plugins' relative paths inside{' '}
						<code>.clothingplugins</code> and <code>.vap</code> files to the package being prepared.
					</p>
				</div>
			</section>

			<div className="actions -left">
				<button
					className={`clear ${config.onTop ? '-active' : ''}`}
					onClick={() => SetConfig({...config, onTop: !config.onTop})}
					title="Toggle window On Top"
				>
					{icons.circleFull}
				</button>
				{messages.length > 0 && (
					<button className="clear" onClick={() => setMessages([])} title="Clear output history">
						Clear
					</button>
				)}
			</div>
		</main>
	);
}

const variantSeverity: Record<string, number> = {
	info: 1,
	warning: 2,
	success: 3,
	danger: 4,
};

function Message({data}: {data: lib.Message}) {
	const hasNotes = data.notes && data.notes.length > 0;
	const title = useMemo(() => `${data.title}`.split('/').at(-1), [data.title]);
	const variant = useMemo(() => {
		let variant: string | null = null;
		let severity = 0;
		for (const note of data.notes) {
			let noteVariantSeverity = variantSeverity[note.variant] || 0;
			if (noteVariantSeverity > severity) {
				severity = noteVariantSeverity;
				variant = note.variant;
			}
		}
		return variant ? `-${variant}` : '';
	}, [data.notes]);

	return (
		<article className={`Message ${variant}`}>
			<hgroup>
				{data.icon && data.icon in icons && icons[data.icon]}
				<h1 title={data.title}>{title}</h1>
			</hgroup>
			{hasNotes && <Notes data={data.notes} />}
		</article>
	);
}

function Notes({data}: {data: lib.Note[]}) {
	return (
		<ul className="Notes">
			{data.map((note, i) => (
				<Note key={i} data={note} />
			))}
		</ul>
	);
}

function Note({data}: {data: lib.Note}) {
	const hasDetails = data.details;
	const [showDetails, setShowDetails] = useState(false);

	return (
		<li className={`-${data.variant} ${showDetails ? '-expanded' : ''}`}>
			<header onClick={hasDetails ? () => setShowDetails(!showDetails) : undefined}>
				{icons[data.variant]}
				<h2
					dangerouslySetInnerHTML={{__html: data.text}}
					title={hasDetails && (showDetails ? 'Hide details' : 'Show details')}
				/>
				{hasDetails && (showDetails ? icons.arrowUp : icons.arrowDown)}
			</header>
			{showDetails && (
				<pre>
					<code>{data.details}</code>
				</pre>
			)}
		</li>
	);
}

const icons: Record<string, JSX.Element> = {
	success: (
		<svg
			className="Icon check-circle"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="m424-408-86-86q-11-11-28-11t-28 11q-11 11-11 28t11 28l114 114q12 12 28 12t28-12l226-226q11-11 11-28t-11-28q-11-11-28-11t-28 11L424-408Zm56 328q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z" />
		</svg>
	),
	danger: (
		<svg
			className="Icon error-circle"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M480-280q17 0 28.5-11.5T520-320q0-17-11.5-28.5T480-360q-17 0-28.5 11.5T440-320q0 17 11.5 28.5T480-280Zm0-160q17 0 28.5-11.5T520-480v-160q0-17-11.5-28.5T480-680q-17 0-28.5 11.5T440-640v160q0 17 11.5 28.5T480-440Zm0 360q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z" />
		</svg>
	),
	warning: (
		<svg
			className="Icon warning-circle"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M109-120q-11 0-20-5.5T75-140q-5-9-5.5-19.5T75-180l370-640q6-10 15.5-15t19.5-5q10 0 19.5 5t15.5 15l370 640q6 10 5.5 20.5T885-140q-5 9-14 14.5t-20 5.5H109Zm69-80h604L480-720 178-200Zm302-40q17 0 28.5-11.5T520-280q0-17-11.5-28.5T480-320q-17 0-28.5 11.5T440-280q0 17 11.5 28.5T480-240Zm0-120q17 0 28.5-11.5T520-400v-120q0-17-11.5-28.5T480-560q-17 0-28.5 11.5T440-520v120q0 17 11.5 28.5T480-360Zm0-100Z" />
		</svg>
	),
	info: (
		<svg
			className="Icon info-circle"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M480-280q17 0 28.5-11.5T520-320v-160q0-17-11.5-28.5T480-520q-17 0-28.5 11.5T440-480v160q0 17 11.5 28.5T480-280Zm0-320q17 0 28.5-11.5T520-640q0-17-11.5-28.5T480-680q-17 0-28.5 11.5T440-640q0 17 11.5 28.5T480-600Zm0 520q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z" />
		</svg>
	),
	arrowDown: (
		<svg
			className="Icon arrow arrow-down"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M480-361q-8 0-15-2.5t-13-8.5L268-556q-11-11-11-28t11-28q11-11 28-11t28 11l156 156 156-156q11-11 28-11t28 11q11 11 11 28t-11 28L508-372q-6 6-13 8.5t-15 2.5Z" />
		</svg>
	),
	arrowUp: (
		<svg
			className="Icon arrow arrow-up"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M480-528 324-372q-11 11-28 11t-28-11q-11-11-11-28t11-28l184-184q12-12 28-12t28 12l184 184q11 11 11 28t-11 28q-11 11-28 11t-28-11L480-528Z" />
		</svg>
	),
	file: (
		<svg
			className="Icon file"
			xmlns="http://www.w3.org/2000/svg"
			height="24px"
			viewBox="0 -960 960 960"
			width="24px"
			fill="currentColor"
		>
			<path d="M360-240h240q17 0 28.5-11.5T640-280q0-17-11.5-28.5T600-320H360q-17 0-28.5 11.5T320-280q0 17 11.5 28.5T360-240Zm0-160h240q17 0 28.5-11.5T640-440q0-17-11.5-28.5T600-480H360q-17 0-28.5 11.5T320-440q0 17 11.5 28.5T360-400ZM240-80q-33 0-56.5-23.5T160-160v-640q0-33 23.5-56.5T240-880h287q16 0 30.5 6t25.5 17l194 194q11 11 17 25.5t6 30.5v447q0 33-23.5 56.5T720-80H240Zm280-560v-160H240v640h480v-440H560q-17 0-28.5-11.5T520-640ZM240-800v200-200 640-640Z" />
		</svg>
	),
	circleFull: (
		<svg className="Icon circle-full" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" fill="currentColor">
			<circle r="50" cx="50" cy="50" />
		</svg>
	),
};

export default App;
