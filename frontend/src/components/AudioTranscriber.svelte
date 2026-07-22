<script>
  let files = $state([]);  // Array of { id, file, name, size, status, progress, transcript, error }
  let status = $state('idle'); // idle | loading-model | transcribing | done
  let modelProgress = $state(0);
  let modelProgressText = $state('');
  let dragOver = $state(false);

  // Model pipeline (cached after first load)
  let transcriber = null;
  let fileIdCounter = 0;

  async function loadModel() {
    if (transcriber) return transcriber;

    status = 'loading-model';
    modelProgressText = 'Đang tải model Whisper (~40MB lần đầu)...';
    modelProgress = 0;

    const { pipeline } = await import('@huggingface/transformers');

    transcriber = await pipeline('automatic-speech-recognition', 'onnx-community/whisper-tiny', {
      dtype: 'q8',
      device: 'wasm',
      progress_callback: (data) => {
        if (data.status === 'progress') {
          modelProgress = Math.round(data.progress);
          modelProgressText = `Đang tải model... ${modelProgress}%`;
        } else if (data.status === 'done') {
          modelProgressText = 'Model đã sẵn sàng!';
        }
      },
    });

    return transcriber;
  }

  async function readFileAsAudioData(audioFile) {
    const arrayBuffer = await audioFile.arrayBuffer();
    const audioContext = new (window.AudioContext || window.webkitAudioContext)({
      sampleRate: 16000,
    });
    const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
    return audioBuffer.getChannelData(0);
  }

  async function transcribeAll() {
    if (files.length === 0) return;

    try {
      const model = await loadModel();
      status = 'transcribing';

      for (let i = 0; i < files.length; i++) {
        const item = files[i];
        if (item.status === 'done') continue; // skip already transcribed

        files[i] = { ...files[i], status: 'transcribing', error: '' };

        try {
          const audioData = await readFileAsAudioData(item.file);
          const result = await model(audioData, {
            chunk_length_s: 30,
            stride_length_s: 5,
            language: null,
            task: 'transcribe',
          });

          files[i] = { ...files[i], status: 'done', transcript: result.text.trim() };
        } catch (err) {
          files[i] = { ...files[i], status: 'error', error: err.message || 'Lỗi xử lý file' };
        }
      }

      status = 'done';
    } catch (err) {
      status = 'idle';
      console.error('Model load error:', err);
    }
  }

  function addFiles(newFiles) {
    const validFiles = [];

    for (const f of newFiles) {
      if (!f.type.startsWith('audio/')) continue;
      if (f.size > 10 * 1024 * 1024) continue;

      validFiles.push({
        id: ++fileIdCounter,
        file: f,
        name: f.name,
        size: f.size,
        status: 'pending',
        progress: 0,
        transcript: '',
        error: '',
      });
    }

    files = [...files, ...validFiles];
  }

  function onFileInput(e) {
    const selected = Array.from(e.target.files || []);
    addFiles(selected);
    e.target.value = ''; // reset input
  }

  function onDrop(e) {
    e.preventDefault();
    dragOver = false;
    const dropped = Array.from(e.dataTransfer?.files || []);
    addFiles(dropped);
  }

  function onDragOver(e) {
    e.preventDefault();
    dragOver = true;
  }

  function onDragLeave() {
    dragOver = false;
  }

  function removeFile(id) {
    files = files.filter(f => f.id !== id);
  }

  function reset() {
    files = [];
    status = 'idle';
    modelProgress = 0;
    modelProgressText = '';
  }

  function copyTranscript(text) {
    navigator.clipboard.writeText(text);
  }

  function copyAll() {
    const allText = files
      .filter(f => f.transcript)
      .map(f => `[${f.name}]\n${f.transcript}`)
      .join('\n\n');
    navigator.clipboard.writeText(allText);
  }

  let isProcessing = $derived(status === 'loading-model' || status === 'transcribing');
  let hasResults = $derived(files.some(f => f.transcript));
  let pendingCount = $derived(files.filter(f => f.status === 'pending').length);
</script>

<div class="space-y-6">
  <!-- Upload Area -->
  <div
    class="card p-8 text-center transition-all duration-200 {dragOver ? 'border-indigo-400 dark:border-indigo-500 bg-indigo-50 dark:bg-indigo-950/30 scale-[1.01]' : ''}"
    role="region"
    aria-label="Khu vực upload file audio"
    ondrop={onDrop}
    ondragover={onDragOver}
    ondragleave={onDragLeave}
  >
    <div class="space-y-4">
      <div class="mx-auto w-16 h-16 rounded-full bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center">
        <svg class="w-8 h-8 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
        </svg>
      </div>
      <div>
        <p class="text-lg font-medium text-gray-700 dark:text-gray-300">
          Kéo thả file MP3 vào đây
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Hỗ trợ nhiều file • MP3, WAV, OGG • Mỗi file tối đa 60s / 10MB
        </p>
      </div>
      <label class="btn btn-primary cursor-pointer">
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
        </svg>
        Chọn file audio
        <input
          type="file"
          accept="audio/*"
          multiple
          class="hidden"
          onchange={onFileInput}
        />
      </label>
    </div>
  </div>

  <!-- File List -->
  {#if files.length > 0}
    <div class="card divide-y divide-gray-200 dark:divide-gray-800">
      <div class="p-4 flex items-center justify-between">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {files.length} file{files.length > 1 ? 's' : ''}
          {#if pendingCount > 0}
            <span class="text-gray-500">• {pendingCount} chờ xử lý</span>
          {/if}
        </span>
        <div class="flex gap-2">
          <button
            class="btn btn-primary text-sm"
            onclick={transcribeAll}
            disabled={isProcessing || pendingCount === 0}
          >
            {#if isProcessing}
              <svg class="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Đang xử lý...
            {:else}
              Chuyển thành text
            {/if}
          </button>
          <button
            class="btn btn-secondary text-sm"
            onclick={reset}
            disabled={isProcessing}
          >
            Xoá tất cả
          </button>
        </div>
      </div>

      {#each files as item (item.id)}
        <div class="p-4 space-y-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-8 h-8 shrink-0 rounded-lg flex items-center justify-center
                {item.status === 'done' ? 'bg-green-100 dark:bg-green-900/40' : item.status === 'error' ? 'bg-red-100 dark:bg-red-900/40' : item.status === 'transcribing' ? 'bg-yellow-100 dark:bg-yellow-900/40' : 'bg-gray-100 dark:bg-gray-800'}">
                {#if item.status === 'done'}
                  <svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                {:else if item.status === 'error'}
                  <svg class="w-4 h-4 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                {:else if item.status === 'transcribing'}
                  <svg class="w-4 h-4 text-yellow-600 dark:text-yellow-400 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                {:else}
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/>
                  </svg>
                {/if}
              </div>
              <div class="min-w-0">
                <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{item.name}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{(item.size / 1024).toFixed(0)} KB</p>
              </div>
            </div>
            <button
              class="p-1 text-gray-400 hover:text-red-500 transition-colors"
              onclick={() => removeFile(item.id)}
              disabled={isProcessing}
              aria-label="Xoá file"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
              </svg>
            </button>
          </div>

          {#if item.error}
            <p class="text-xs text-red-600 dark:text-red-400 pl-11">{item.error}</p>
          {/if}

          {#if item.transcript}
            <div class="ml-11 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
              <div class="flex items-start justify-between gap-2">
                <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">{item.transcript}</p>
                <button
                  class="shrink-0 p-1 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                  onclick={() => copyTranscript(item.transcript)}
                  aria-label="Copy transcript"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Model Loading Progress -->
  {#if status === 'loading-model'}
    <div class="card p-4 space-y-2">
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600 dark:text-gray-400">{modelProgressText}</span>
        {#if modelProgress > 0}
          <span class="font-mono text-indigo-600 dark:text-indigo-400">{modelProgress}%</span>
        {/if}
      </div>
      <div class="w-full h-2 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
        <div
          class="h-full bg-indigo-600 dark:bg-indigo-500 rounded-full transition-all duration-300"
          style="width: {modelProgress}%"
        ></div>
      </div>
      <p class="text-xs text-gray-500 dark:text-gray-400">
        Model chỉ cần tải 1 lần, các lần sau sẽ dùng từ cache.
      </p>
    </div>
  {/if}

  <!-- Copy All Button -->
  {#if hasResults}
    <div class="flex justify-end">
      <button class="btn btn-secondary" onclick={copyAll}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
        Copy tất cả kết quả
      </button>
    </div>
  {/if}
</div>
