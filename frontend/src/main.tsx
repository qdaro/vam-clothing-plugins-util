import React from 'react';
import {createRoot} from 'react-dom/client';
import './style.css';
import App from './App';
import {WailsDropInterface} from './lib/wails-drop-interface';

const container = document.getElementById('root');

const root = createRoot(container!);

root.render(
	<React.StrictMode>
		<WailsDropInterface>
			<App />
		</WailsDropInterface>
	</React.StrictMode>
);
