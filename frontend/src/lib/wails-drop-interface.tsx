import {useMemo, useEffect, useRef, useContext, createContext, RefObject, PropsWithChildren} from 'react';
import * as runtime from '../../wailsjs/runtime';
import {getPointToRectProximity} from './utils';

type DropCallback = (paths: string[]) => void;
type DropMap = Map<RefObject<HTMLElement>, RefObject<DropCallback>>;

const WailsDropInterfaceContext = createContext<DropMap | null>(null);

export function WailsDropInterface({children}: PropsWithChildren) {
	const dropMap = useMemo(() => new Map() satisfies DropMap, []);

	useEffect(() => {
		runtime.OnFileDrop((x, y, paths) => {
			for (const [ref, callback] of dropMap.entries()) {
				const rect = ref.current?.getBoundingClientRect();
				const cb = callback.current;
				if (rect && cb && getPointToRectProximity(x, y, rect) === 0) {
					cb(paths);
					break;
				}
			}
		}, true);

		return () => {
			runtime.OnFileDropOff();
		};
	}, []);

	return <WailsDropInterfaceContext.Provider value={dropMap}>{children}</WailsDropInterfaceContext.Provider>;
}

export function useWailsFileDrop(ref: RefObject<HTMLElement>, callback: (paths: string[]) => void) {
	const callbackRef = useRef<((paths: string[]) => void) | null>(null);
	const dropMap = useContext(WailsDropInterfaceContext);
	callbackRef.current = callback;
	useEffect(() => {
		if (!dropMap) return;
		dropMap.set(ref, callbackRef);
		return () => {
			dropMap.delete(ref);
		};
	}, [dropMap]);
}
