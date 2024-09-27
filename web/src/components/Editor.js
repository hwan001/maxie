import React, { useState, useRef, useEffect } from 'react';
import '../styles/Editor.css';

const Editor = ({ data }) => {
  //const [width, setWidth] = useState(800);
  const [height, setHeight] = useState(300);
  const editorRef = useRef(null);
  const isResizing = useRef(false);

  const startResize = (e) => {
    e.preventDefault();
    isResizing.current = true;
  };

  const stopResize = () => {
    isResizing.current = false;
  };

  const resize = (e) => {
    if (!isResizing.current) return;
    // setWidth(e.clientX - editorRef.current.offsetLeft);
    setHeight(e.clientY - editorRef.current.offsetTop);
  };

  useEffect(() => {
    window.addEventListener('mousemove', resize);
    window.addEventListener('mouseup', stopResize);
    return () => {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopResize);
    };
  }, []);

  return (
    <div
      className="editor-container"
      ref={editorRef}
      style={{ height: `${height}px` }} /*width: `${width}px`, */
    >
      {data}
      <div
        className="resize-handle"
        onMouseDown={startResize}
      ></div>
    </div>
  );
};

export default Editor;
