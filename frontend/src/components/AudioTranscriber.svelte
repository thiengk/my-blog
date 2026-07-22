<script>
  let file = $state(null);
  let fileName = $state('');
  let status = $state('idle'); // idle | loading-model | transcribing | done | error
  let progress = $state(0);
  let progressText = $state('');
  let transcript = $state('');
  let errorMessage = $state('');
  let dragOver = $state(false);

  // Model pipeline (cached after first load)
  let transcriber = null;

  async function loadModel() {
    if (transcriber) return transcriber;

    status = 'loading-model';
    progressText = 'Đang tải model Whisper (~40MB lần đầu)...';
    progress = 0;

    // Dynamic import to avoid SSR/build issues
    const { pipeline } = await import('@xenova/transformers');

    transcriber = await pipeline('automatic-speech-recognition', 'Xenova/whisper-tiny', {
      progress_callback: (data) => {
        if (data.status === 'progress') {
          progress = Math.round(data.progress);
          progressText = `Đang tải model... ${progress}%`;
        } else if (data.status === 'done') {
          progressText = 'Model đã sẵn sàng!';
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

    // Get mono channel (Whisper expects mono 16kHz)
    const audioData = audioBuffer.getChannelData(0);
    return audioData;
  }

  async function transcribe() {
    if (!file) return;

    try {
      errorMessage = '';
      transcript = '';

      const model = await loadModel();

      status = 'transcribing';
      progressText = 'Đang chuyển đổi audio thành text...';
      progress = 0;

      const audioData = await readFileAsAudioData(file);

      const result = await model(audioData, {
        chunk_length_s: 30,
        stride_length_s: 5,
        language: null, // auto-detect language
        task: 'transcribe',
      });

      transcript = result.text.trim();
      status = 'done';
      progressText = '';
    } catch (err) {
      status = 'error';
      errorMessage = err.message || 'Có lỗi xảy ra khi xử lý audio';
      console.error('Transcription error:', err);
    }
  }

  function handleFile(selectedFile) {
    if (!selectedFile) return;

    // Validate file type
    if (!selectedFile.type.startsWith('audio/')) {
      errorMessage = 'Vui lòng chọn file audio (MP3, WAV, OGG, etc.)';
      return;
    }

    // Validate file size (max ~10MB)
    if (selectedFile.size > 10 * 1024 * 1024) {
      errorMessage = 'File quá lớn. Vui lòng chọn file dưới 10MB.';
      return;
    }

    file = selectedFile;
    fileName = selectedFile.name;
    errorMessage = '';
    transcript = '';
    status = 'idle';
  }

  function onFileInput(e) {
    const selectedFile = e.target.files?.[0];
    handleFile(selectedFile);
  }

  function onDrop(e) {
    e.preventDefault();
    dragOver = false;
    const droppedFile = e.dataTransfer?.files?.[0];
    handleFile(droppedFile);
  }

  function onDragOver(e) {
    e.preventDefault();
    dragOver = true;
  }

  function onDragLeave() {
    dragOver = false;
  }

  function reset() {
    file = null;
    fileName = '';
    status = 'idle';
    progress = 0;
    progressText = '';
    transcript = '';
    errorMessage = '';
  }

  function copyTranscript() {
    navigator.clipboard.writeText(transcript);
  }

  let isProcessing = $derived(status === 'loading-model' || status === 'transcribing');
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
    {#if !file}
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
            hoặc click để chọn file • MP3, WAV, OGG • Tối đa 60s / 10MB
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
            class="hidden"
            onchange={onFileInput}
          />
        </label>
      </div>
    {:else}
      <!-- File selected -->
      <div class="space-y-4">
        <div class="flex items-center justify-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center">
            <svg class="w-5 h-5 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"/>
            </svg>
          </div>
          <div class="text-left">
            <p class="font-medium text-gray-900 dark:text-gray-100">{fileName}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {(file.size / 1024).toFixed(0)} KB
            </p>
          </div>
        </div>

        <div class="flex items-center justify-center gap-3">
          <button
            class="btn btn-primary"
            onclick={transcribe}
            disabled={isProcessing}
          >
            {#if isProcessing}
              <svg class="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Đang xử lý...
            {:else}
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
              </svg>
              Chuyển thành text
            {/if}
          </button>
          <button
            class="btn btn-secondary"
            onclick={reset}
            disabled={isProcessing}
          >
            Đổi file
          </button>
        </div>
      </div>
    {/if}
  </div>

  <!-- Progress -->
  {#if isProcessing}
    <div class="card p-4 space-y-2">
      <div class="flex items-center justify-between text-sm">
        <span class="text-gray-600 dark:text-gray-400">{progressText}</span>
        {#if progress > 0 && status === 'loading-model'}
          <span class="font-mono text-indigo-600 dark:text-indigo-400">{progress}%</span>
        {/if}
      </div>
      <div class="w-full h-2 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
        <div
          class="h-full bg-indigo-600 dark:bg-indigo-500 rounded-full transition-all duration-300 {status === 'transcribing' ? 'animate-pulse' : ''}"
          style="width: {status === 'loading-model' ? progress : 100}%"
        ></div>
      </div>
      {#if status === 'loading-model'}
        <p class="text-xs text-gray-500 dark:text-gray-400">
          Model chỉ cần tải 1 lần, các lần sau sẽ dùng từ cache.
        </p>
      {/if}
    </div>
  {/if}

  <!-- Error -->
  {#if errorMessage}
    <div class="card border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/30 p-4">
      <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-red-500 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <p class="text-sm text-red-700 dark:text-red-300">{errorMessage}</p>
      </div>
    </div>
  {/if}

  <!-- Result -->
  {#if transcript}
    <div class="card p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          📝 Kết quả
        </h2>
        <button
          class="btn btn-secondary text-xs"
          onclick={copyTranscript}
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
          </svg>
          Copy
        </button>
      </div>
      <div class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
        <p class="text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">{transcript}</p>
      </div>
    </div>
  {/if}
</div>
