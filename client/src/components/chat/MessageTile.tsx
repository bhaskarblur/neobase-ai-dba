import { AlertCircle, Braces, Clock, Copy, History, Loader, Pencil, Play, Send, Table, X, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { useStream } from '../../contexts/StreamContext';
import chatService from '../../services/chatService';
import ConfirmationModal from '../modals/ConfirmationModal';
import RollbackConfirmationModal from '../modals/RollbackConfirmationModal';
import LoadingSteps from './LoadingSteps';
import { Message, QueryResult } from './types';

interface QueryState {
    isExecuting: boolean;
    isExample: boolean;
}

interface MessageTileProps {
    chatId: string;
    message: Message;
    onEdit?: (id: string) => void;
    editingMessageId: string | null;
    editInput: string;
    setEditInput: (input: string) => void;
    onSaveEdit: (id: string, content: string) => void;
    onCancelEdit: () => void;
    queryStates: Record<string, QueryState>;
    setQueryStates: React.Dispatch<React.SetStateAction<Record<string, QueryState>>>;
    queryTimeouts: React.MutableRefObject<Record<string, NodeJS.Timeout>>;
}

const toastStyle = {
    style: {
        background: '#000',
        color: '#fff',
        border: '4px solid #000',
        borderRadius: '12px',
        boxShadow: '4px 4px 0px 0px rgba(0,0,0,1)',
        padding: '12px 24px',
        fontSize: '14px',
        fontWeight: '500',
    },
    position: 'bottom-center' as const,
    duration: 2000,
};

export default function MessageTile({
    chatId,
    message,
    onEdit,
    editingMessageId,
    editInput,
    setEditInput,
    onSaveEdit,
    onCancelEdit,
    queryStates,
    setQueryStates,
    queryTimeouts,
}: MessageTileProps) {
    const { streamId } = useStream();
    const [viewMode, setViewMode] = useState<'table' | 'json'>('table');
    const [showCriticalConfirm, setShowCriticalConfirm] = useState(false);
    const [queryToExecute, setQueryToExecute] = useState<string | null>(null);
    const [rollbackState, setRollbackState] = useState<{
        show: boolean;
        queryId: string | null;
    }>({ show: false, queryId: null });
    const [streamingQueryIndex, setStreamingQueryIndex] = useState<number>(-1);
    const [isDescriptionStreaming, setIsDescriptionStreaming] = useState(false);
    const [isQueryStreaming, setIsQueryStreaming] = useState(false);
    const [currentDescription, setCurrentDescription] = useState('');
    const [currentQuery, setCurrentQuery] = useState('');

    useEffect(() => {
        const streamQueries = async () => {
            if (!message.queries || !message.is_streaming) return;

            // Stream each query one by one
            for (let i = 0; i < message.queries.length; i++) {
                const query = message.queries[i];
                setStreamingQueryIndex(i);

                // Stream description
                setIsDescriptionStreaming(true);
                for (let j = 0; j <= query.description.length; j++) {
                    await new Promise(resolve => setTimeout(resolve, 40));
                    setCurrentDescription(query.description.slice(0, j));
                }
                setIsDescriptionStreaming(false);

                // Only stream query text if message is not currently streaming
                if (!message.is_streaming) {
                    setIsQueryStreaming(true);
                    for (let j = 0; j <= query.query.length; j++) {
                        await new Promise(resolve => setTimeout(resolve, 40));
                        setCurrentQuery(query.query.slice(0, j));
                    }
                    setIsQueryStreaming(false);
                }

                // Mark query as done streaming
                if (message.queries) {
                    message.queries[i].is_streaming = false;
                }
            }
            setStreamingQueryIndex(-1);
        };

        streamQueries();
    }, [message.queries, message.is_streaming]);

    const handleCopyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast.success('Copied to clipboard!', toastStyle);
    };

    const handleExecuteQuery = async (queryId: string) => {
        const query = message.queries?.find(q => q.id === queryId);
        if (!query) return;

        if (query.is_critical) {
            setQueryToExecute(queryId);
            setShowCriticalConfirm(true);
            return;
        }
        executeQuery(queryId);
    };

    const executeQuery = async (queryId: string) => {
        const query = message.queries?.find(q => q.id === queryId);
        if (!query) return;

        if (queryTimeouts.current[queryId]) {
            clearTimeout(queryTimeouts.current[queryId]);
            delete queryTimeouts.current[queryId];
        }

        setQueryStates(prev => ({
            ...prev,
            [queryId]: { isExecuting: true, isExample: false }
        }));

        try {
            await chatService.executeQuery(chatId, message.id, query.id, streamId || '');

            // Update query state
            if (message.queries) {
                message.queries.find(q => q.id === queryId)!.is_executed = true;
                message.queries.find(q => q.id === queryId)!.is_rolled_back = false;
            }
        } catch (error: any) {
            console.log('error', error.message);
            toast.error("Query execution failed");
        } finally {
            setQueryStates(prev => ({
                ...prev,
                [queryId]: { isExecuting: false, isExample: false }
            }));
        }
    };

    const handleRollback = async (queryId: string) => {
        const queryIndex = message.queries?.findIndex(q => q.id === queryId) ?? -1;
        if (queryIndex === -1) return;


        try {
            setQueryStates(prev => ({
                ...prev,
                [queryId]: { isExecuting: true, isExample: true }
            }));

            await chatService.rollbackQuery(chatId, message.id, queryId, streamId || '');

            // Update query state
            if (message.queries) {
                message.queries[queryIndex].is_rolled_back = true;
            }

        } catch (error: any) {
            toast.error(error.message);
        } finally {
            setQueryStates(prev => ({
                ...prev,
                [queryId]: { isExecuting: false, isExample: true }
            }));
            setRollbackState({ show: false, queryId: null });
        }
    };

    const renderTableView = (data: any[]) => {
        if (!data.length) return null;
        const columns = Object.keys(data[0]);

        return (
            <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr>
                            {columns.map(column => (
                                <th key={column} className="py-2 px-4 bg-gray-800 border-b border-gray-700 text-gray-300 font-mono">
                                    {column}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {data.map((row, i) => (
                            <tr key={i} className="border-b border-gray-700">
                                {columns.map(column => (
                                    <td key={column} className="py-2 px-4">
                                        <span className={`${typeof row[column] === 'number'
                                            ? 'text-cyan-400'
                                            : typeof row[column] === 'boolean'
                                                ? 'text-purple-400'
                                                : column.includes('time') || column.includes('date')
                                                    ? 'text-yellow-300'
                                                    : 'text-green-400'
                                            }`}>
                                            {JSON.stringify(row[column])}
                                        </span>
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        );
    };

    const renderQuery = (isMessageStreaming: boolean, query: QueryResult, index: number) => {
        const queryId = query.id;
        const shouldShowExampleResult = !query.is_executed && !query.is_rolled_back;
        const resultToShow = shouldShowExampleResult ? query.example_result : query.execution_result;
        const isCurrentlyStreaming = !isMessageStreaming && streamingQueryIndex === index;

        const shouldShowRollback = query.can_rollback &&
            query.is_executed &&
            !query.is_rolled_back;

        return (
            <div>
                <p className="mb-4 mt-4 font-base text-base">
                    <span className="text-black font-semibold">Explanation:</span> {isCurrentlyStreaming && isDescriptionStreaming
                        ? currentDescription
                        : query.description}
                </p>
                <div key={index} className="mt-4 bg-black text-white rounded-lg font-mono text-sm overflow-hidden w-full" style={{ minWidth: '100%' }}>
                    <div className="flex flex-wrap items-center justify-between gap-2 mb-4 px-4 pt-4">

                        <div className="flex items-center gap-2">
                            <span>Query {index + 1}:</span>
                        </div>
                        <div className="flex items-center">
                            {queryStates[queryId]?.isExecuting ? (
                                <button
                                    onClick={(e) => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        if (queryTimeouts.current[queryId]) {
                                            clearTimeout(queryTimeouts.current[queryId]);
                                            delete queryTimeouts.current[queryId];
                                        }
                                        setQueryStates((prev: Record<string, QueryState>) => ({
                                            ...prev,
                                            [queryId]: { isExecuting: false, isExample: true }
                                        }));
                                        setTimeout(() => {
                                            window.scrollTo(window.scrollX, window.scrollY);
                                        }, 0);
                                        toast.error('Query cancelled', toastStyle);
                                    }}
                                    className="p-2 hover:bg-gray-800 rounded transition-colors text-red-500 hover:text-red-400"
                                    title="Cancel query"
                                >
                                    <XCircle className="w-4 h-4" />
                                </button>
                            ) : (
                                <button
                                    onClick={(e) => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        handleExecuteQuery(queryId);
                                        setTimeout(() => {
                                            window.scrollTo(window.scrollX, window.scrollY);
                                        }, 0);
                                    }}
                                    className="p-2 hover:bg-gray-800 rounded transition-colors text-red-500 hover:text-red-400"
                                    title="Run query"
                                >
                                    <Play className="w-4 h-4" />
                                </button>
                            )}
                            <div className="w-px h-4 bg-gray-700 mx-2" />
                            <button
                                onClick={() => handleCopyToClipboard(query.query)}
                                className="p-2 hover:bg-gray-800 rounded transition-colors text-white hover:text-gray-200"
                                title="Copy query"
                            >
                                <Copy className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                    <pre className={`
                    text-sm overflow-x-auto p-4 border-t border-gray-700
                    ${isCurrentlyStreaming && isQueryStreaming ? 'animate-pulse duration-300' : ''}
                `}>
                        <code className="whitespace-pre-wrap break-words">
                            {isCurrentlyStreaming && isQueryStreaming
                                ? currentQuery
                                : query.query}
                        </code>
                    </pre>
                    {(query.execution_result || query.example_result || query.error) && (
                        <div className="border-t border-gray-700 mt-2 w-full">
                            {queryStates[queryId]?.isExecuting ? (
                                <div className="flex items-center justify-center p-8">
                                    <Loader className="w-8 h-8 animate-spin text-gray-400" />
                                    <span className="ml-3 text-gray-400">Executing query...</span>
                                </div>
                            ) : (
                                <div className="mt-3 px-4 pt-4 w-full">
                                    <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
                                        <div className="flex items-center gap-2 text-gray-400">
                                            {query.error ? (
                                                <span className="text-neo-error font-medium flex items-center gap-2">
                                                    <AlertCircle className="w-4 h-4" />
                                                    Error
                                                </span>
                                            ) : (
                                                <span>
                                                    {shouldShowExampleResult ? 'Example Result:' : 'Result:'}
                                                </span>
                                            )}
                                            {query.example_execution_time && !query.execution_time && !query.is_executed && (
                                                <span className="text-xs bg-gray-800 px-2 py-1 rounded flex items-center gap-1">
                                                    <Clock className="w-3 h-3" />
                                                    {query.example_execution_time.toLocaleString()}ms
                                                </span>
                                            )}
                                            {query.execution_time && (
                                                <span className="text-xs bg-gray-800 px-2 py-1 rounded flex items-center gap-1">
                                                    <Clock className="w-3 h-3" />
                                                    {query.execution_time.toLocaleString()}ms
                                                </span>
                                            )}
                                        </div>
                                        {!query.error && <div className="flex gap-2">
                                            <div className="flex items-center">
                                                <button
                                                    onClick={(e) => {
                                                        e.preventDefault();
                                                        e.stopPropagation();
                                                        setViewMode('table');
                                                        setTimeout(() => {
                                                            window.scrollTo(window.scrollX, window.scrollY);
                                                        }, 0);
                                                    }}
                                                    className={`p-1 md:p-2 rounded ${viewMode === 'table' ? 'bg-gray-700' : 'hover:bg-gray-800'}`}
                                                    title="Table view"
                                                >
                                                    <Table className="w-4 h-4" />
                                                </button>
                                                <div className="w-px h-4 bg-gray-700 mx-2" />
                                                <button
                                                    onClick={(e) => {
                                                        e.preventDefault();
                                                        e.stopPropagation();
                                                        setViewMode('json');
                                                        setTimeout(() => {
                                                            window.scrollTo(window.scrollX, window.scrollY);
                                                        }, 0);
                                                    }}
                                                    className={`p-1 md:p-2 rounded ${viewMode === 'json' ? 'bg-gray-700' : 'hover:bg-gray-800'}`}
                                                    title="JSON view"
                                                >
                                                    <Braces className="w-4 h-4" />
                                                </button>
                                                <div className="w-px h-4 bg-gray-700 mx-2" />
                                                <button
                                                    onClick={() => handleCopyToClipboard(JSON.stringify(resultToShow, null, 2))}
                                                    className="p-2 hover:bg-gray-800 rounded text-white hover:text-gray-200"
                                                    title="Copy result"
                                                >
                                                    <Copy className="w-4 h-4" />
                                                </button>
                                                {shouldShowRollback && (
                                                    <button
                                                        onClick={(e) => {
                                                            e.preventDefault();
                                                            e.stopPropagation();
                                                            setRollbackState({ show: true, queryId });
                                                            setTimeout(() => {
                                                                window.scrollTo(window.scrollX, window.scrollY);
                                                            }, 0);
                                                            handleRollback(query.id);
                                                        }}
                                                        className="p-2 hover:bg-gray-800 rounded text-yellow-400 hover:text-yellow-300"
                                                        disabled={queryStates[queryId]?.isExecuting}
                                                    >
                                                        <History className="w-4 h-4" />
                                                    </button>
                                                )}
                                            </div>
                                        </div>}
                                    </div>
                                    {query.error ? (
                                        <div className="bg-neo-error/10 text-neo-error p-4 rounded-lg mb-6">
                                            <div className="font-bold mb-2">{query.error.code}</div>
                                            {query.error.message != query.error.details && <div className="mb-2">{query.error.message}</div>}
                                            {query.error.details && (
                                                <div className="text-sm opacity-80 border-t border-neo-error/20 pt-2 mt-2">
                                                    {query.error.details}
                                                </div>
                                            )}
                                        </div>
                                    ) : (
                                        <div className="w-full">
                                            <div className={`
                                            text-green-400 pb-6 w-full
                                            ${!query.example_result && !query.error ? 'animate-pulse duration-300' : ''}
                                        `}>
                                                {viewMode === 'table' ? (
                                                    <div className="w-full">
                                                        {Object.keys(resultToShow || {}).length === 0 ? (
                                                            <div className="w-full">
                                                                Query Successful
                                                            </div>
                                                        ) : (
                                                            <div className="w-full">
                                                                {renderTableView(resultToShow || [])}
                                                            </div>
                                                        )}
                                                    </div>
                                                ) : (
                                                    <div className="w-full">
                                                        <pre className="overflow-x-auto whitespace-pre-wrap">
                                                            {JSON.stringify(resultToShow, null, 2)}
                                                        </pre>
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        );
    };

    return (
        <div className="py-4 md:py-6 first:pt-8 w-full">
            <div className={`
        group flex items-center relative
        ${message.type === 'user' ? 'justify-end' : 'justify-start'}
        w-full
      `}>
                {message.type === 'user' && (
                    <div className="
            absolute 
            right-0 
            -bottom-9
            md:-bottom-10 
            flex 
            gap-1
            z-[5]

          ">
                        <button
                            onClick={() => handleCopyToClipboard(message.content)}
                            className="
                -translate-y-1/2
                p-1.5
                md:p-2 
                group-hover:opacity-100 
                transition-colors
                hover:bg-neo-gray
                rounded-lg
                flex-shrink-0
                border-0
                bg-white/80
                backdrop-blur-sm
              "
                            title="Copy message"
                        >
                            <Copy className="w-4 h-4 text-gray-800" />
                        </button>
                        {onEdit && (
                            <button
                                onClick={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    onEdit(message.id);
                                    setTimeout(() => {
                                        window.scrollTo(window.scrollX, window.scrollY);
                                    }, 0);
                                }}
                                className="
                  -translate-y-1/2
                  p-1.5
                  md:p-2
                  group-hover:opacity-100 
                  hover:bg-neo-gray
                  transition-colors
                  rounded-lg
                  flex-shrink-0
                  border-0
                  bg-white/80
                  backdrop-blur-sm

                "
                                title="Edit message"
                            >
                                <Pencil className="w-4 h-4 text-gray-800" />
                            </button>
                        )}
                    </div>
                )}
                <div className={`
    message-bubble
    inline-block
    ${message.type === 'user' ? (
                        editingMessageId === message.id
                            ? 'w-[95%] sm:w-[85%] md:w-[75%]'
                            : 'w-fit max-w-[95%] sm:max-w-[85%] md:max-w-[75%]'
                    ) : 'w-fit max-w-[95%] sm:max-w-[85%] md:max-w-[75%]'}
    ${message.type === 'user'
                        ? 'message-bubble-user'
                        : 'message-bubble-ai'
                    }
`}>
                    <div className={`
        ${editingMessageId === message.id ? 'w-full min-w-full' : 'w-auto min-w-0'}
        ${message.queries?.length ? 'min-w-full' : ''}
    `}>
                        <div className="relative">
                            {message.content.length === 0 && message.loading_steps && message.loading_steps.length > 0 && (
                                <div className={`
                                    ${message.content ? 'animate-fade-up-out absolute w-full' : ''}
                                    text-gray-700
                                `}>
                                    <LoadingSteps
                                        steps={message.loading_steps.map((step, index) => ({
                                            text: step.text,
                                            done: index !== message.loading_steps!.length - 1
                                        }))}
                                    />
                                </div>
                            )}

                            {editingMessageId === message.id ? (
                                <div className='w-full'>
                                    <textarea
                                        value={editInput}
                                        onChange={(e) => {
                                            e.preventDefault();
                                            e.stopPropagation();
                                            setEditInput(e.target.value);
                                            setTimeout(() => {
                                                window.scrollTo(window.scrollX, window.scrollY);
                                            }, 0);
                                        }}
                                        className="neo-input w-full text-lg min-h-[42px] resize-y py-2 px-3 leading-normal whitespace-pre-wrap"
                                        rows={Math.min(
                                            Math.max(
                                                editInput.split('\n').length,
                                                Math.ceil(editInput.length / 50)
                                            ),
                                            10
                                        )}
                                        autoFocus
                                    />
                                    <div className="flex gap-2 mt-3">
                                        <button
                                            onClick={() => {
                                                onCancelEdit();
                                                setTimeout(() => {
                                                    window.scrollTo(window.scrollX, window.scrollY);
                                                }, 0);
                                            }}
                                            className="neo-button-secondary flex-1 flex items-center justify-center gap-2"
                                        >
                                            <X className="w-4 h-4" />
                                            <span>Cancel</span>
                                        </button>
                                        <button
                                            onClick={() => onSaveEdit(message.id, editInput)}
                                            className="neo-button flex-1 flex items-center justify-center gap-2"
                                        >
                                            <Send className="w-4 h-4" />
                                            <span>Send</span>
                                        </button>
                                    </div>
                                </div>
                            ) : (
                                <div className={message.loading_steps ? 'animate-fade-in' : ''}>
                                    <p className="text-lg whitespace-pre-wrap break-words">{message.content}</p>
                                    {message.queries && message.queries.length > 0 && (
                                        <div className="min-w-full">
                                            {message.queries.map((query: QueryResult, index: number) =>
                                                renderQuery(message.is_streaming || false, query, index)
                                            )}
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {rollbackState.show && rollbackState.queryId && (
                <RollbackConfirmationModal
                    onConfirm={() => handleRollback(rollbackState.queryId!)}
                    onCancel={() => setRollbackState({ show: false, queryId: null })}
                />
            )}

            {showCriticalConfirm && (
                <ConfirmationModal
                    title="Critical Query"
                    message="This query may affect important data. Are you sure you want to proceed?"
                    onConfirm={() => {
                        setShowCriticalConfirm(false);
                        if (queryToExecute !== null) {
                            executeQuery(queryToExecute);
                            setQueryToExecute(null);
                        }
                    }}
                    onCancel={() => {
                        setShowCriticalConfirm(false);
                        setQueryToExecute(null);
                    }}
                />
            )}
        </div>
    );
}