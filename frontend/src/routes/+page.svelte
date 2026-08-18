<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, fly, slide } from 'svelte/transition';

  // API base URL
  const API_URL = 'http://localhost:8081';

  // Types based on Go struct casing
  interface Ticket {
    ticket_id: string;
    ticket_type: string;
    subject: string;
    site?: string;
    site_group?: string;
    region_site?: string;
    domain_group: string;
    company: string;
    country: string;
    priority: string;
    ticket_status: string;
    customer_name: string;
    ProductT1?: string;
    ProductT2?: string;
    ProductT3?: string;
    GroupAssignee?: string;
    Assignee?: string;
    CreatedDate: string;
    ResolvedAt?: string;
    ClosedDate?: string;
    CompleteTime?: string;
    DetailDescription?: string;
    SourceFile: string;
    uploaded_at: string;
    related_type?: string;
  }

  // Reactive state variables using Svelte 5 Runes
  let isLoading = $state(true);
  let tickets = $state<Ticket[]>([]);
  let months = $state<string[]>([]);
  
  // Filtering and view states
  let selectedMonth = $state<string>('');
  let selectedPriority = $state<string>('All');
  let searchQuery = $state<string>('');
  let selectedTicket = $state<Ticket | null>(null); // For detail view drawer
  let showUploadModal = $state(false);

  // Category Filtering States
  let selectedT1 = $state('All T1');
  let selectedT2 = $state('All T2');
  let selectedT3 = $state('All T3');

  // Sorting
  let sortField = $state<keyof Ticket>('CreatedDate');
  let sortDirection = $state<'asc' | 'desc'>('desc');

  // Upload States
  let isDragging = $state(false);
  let uploadState = $state<'idle' | 'selected' | 'uploading' | 'success' | 'error'>('idle');
  let selectedFile = $state<File | null>(null);
  let uploadError = $state('');
  let uploadStats = $state({ inserted: 0, skipped: 0, filename: '' });

  // Pagination
  let currentPage = $state(1);
  const itemsPerPage = 8;

  let isInitialized = $state(false);

  // Initialize and load data
  onMount(async () => {
    await initializeData();
  });

  async function initializeData() {
    isLoading = true;
    try {
      // Get months list first
      const monthsRes = await fetch(`${API_URL}/months`);
      if (monthsRes.ok) {
        months = await monthsRes.json() || [];
        if (months.length > 0 && !selectedMonth) {
          selectedMonth = months[0]; // default to latest month
        }
      }
    } catch (e) {
      console.error('Failed to connect to backend api:', e);
    } finally {
      isInitialized = true;
      isLoading = false;
    }
  }

  async function fetchTickets() {
    try {
      const url = selectedMonth 
        ? `${API_URL}/tickets?month=${selectedMonth}`
        : `${API_URL}/tickets`;
      
      const res = await fetch(url);
      if (res.ok) {
        tickets = await res.json() || [];
        currentPage = 1; // reset pagination on filter change
      }
    } catch (e) {
      console.error('Error fetching tickets:', e);
    }
  }

  // --- Description Filter Engine ---
  let descriptionSearchInput = $state<string>('');
  let isMonthDropdownOpen = $state(false);
  // --- End of Description Filter Engine ---

  // Trigger ticket fetch when selectedMonth changes, only after initial month lookup
  $effect(() => {
    if (isInitialized && selectedMonth !== undefined) {
      fetchTickets();
    }
  });

  // Client-side filtering & sorting using derived state
  const filteredAndSortedTickets = $derived.by(() => {
    let result = [...tickets];

    // Description Search (Regex) Filter
    if (descriptionSearchInput.trim()) {
      try {
        const regex = new RegExp(descriptionSearchInput.trim(), 'i');
        result = result.filter(t => regex.test(t.subject || '') || regex.test(t.DetailDescription || ''));
      } catch (e) {
        // Silently ignore invalid regex during typing
      }
    }

    // Priority filter
    if (selectedPriority !== 'All') {
      result = result.filter(t => t.priority.toLowerCase() === selectedPriority.toLowerCase());
    }

    // Category T1 filter
    if (selectedT1 !== 'All T1') {
      result = result.filter(t => t.ProductT1 === selectedT1);
    }

    // Category T2 filter
    if (selectedT2 !== 'All T2') {
      result = result.filter(t => t.ProductT2 === selectedT2);
    }

    // Category T3 filter
    if (selectedT3 !== 'All T3') {
      result = result.filter(t => t.ProductT3 === selectedT3);
    }

    // Search query filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(t => 
        t.ticket_id.toLowerCase().includes(query) ||
        t.subject.toLowerCase().includes(query) ||
        (t.Assignee && t.Assignee.toLowerCase().includes(query)) ||
        (t.customer_name && t.customer_name.toLowerCase().includes(query)) ||
        t.company.toLowerCase().includes(query)
      );
    }

    // Sorting
    result.sort((a, b) => {
      let valA = a[sortField] || '';
      let valB = b[sortField] || '';

      if (sortField === 'CreatedDate' || sortField === 'ResolvedAt' || sortField === 'ClosedDate') {
        const timeA = valA ? new Date(valA as string).getTime() : 0;
        const timeB = valB ? new Date(valB as string).getTime() : 0;
        return sortDirection === 'asc' ? timeA - timeB : timeB - timeA;
      }

      const strA = String(valA).toLowerCase();
      const strB = String(valB).toLowerCase();
      if (strA < strB) return sortDirection === 'asc' ? -1 : 1;
      if (strA > strB) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });

    return result;
  });

  // Pagination helper
  const paginatedTickets = $derived(
    filteredAndSortedTickets.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage)
  );

  const totalPages = $derived(
    Math.ceil(filteredAndSortedTickets.length / itemsPerPage) || 1
  );

  // Metrics Calculations using derived state
  const metrics = $derived.by(() => {
    let list = tickets;

    // Apply Description Search (Regex) Filter
    if (descriptionSearchInput.trim()) {
      try {
        const regex = new RegExp(descriptionSearchInput.trim(), 'i');
        list = list.filter(t => regex.test(t.subject || '') || regex.test(t.DetailDescription || ''));
      } catch (e) {}
    }

    const total = list.length;
    if (total === 0) {
      return { total: 0, totalWithoutIncident: 0, avgTicketsPerNonIncidentDay: '0.0', critical: 0, high: 0, resolvedRate: 0, avgHrs: '0.0' };
    }

    // Group tickets in selected month by day
    const days: Record<number, number> = {};
    list.forEach(t => {
      const date = new Date(t.CreatedDate);
      if (!isNaN(date.getTime())) {
        const day = date.getDate();
        days[day] = (days[day] || 0) + 1;
      }
    });

    const isIncidentTicket = (t: Ticket) => {
      if (!t.related_type) return false;
      const rt = t.related_type.toLowerCase();
      return rt === 'parent' || rt === 'child';
    };

    const nonIncidentTickets = list.filter(t => !isIncidentTicket(t));
    const totalWithoutIncident = nonIncidentTickets.length;

    // Group non-incident tickets by day to find active normal days
    const nonIncidentDays: Record<number, number> = {};
    nonIncidentTickets.forEach(t => {
      const date = new Date(t.CreatedDate);
      if (!isNaN(date.getTime())) {
        const day = date.getDate();
        nonIncidentDays[day] = (nonIncidentDays[day] || 0) + 1;
      }
    });

    const nonIncidentDaysCount = Object.keys(nonIncidentDays).length;
    const avgTicketsPerNonIncidentDay = nonIncidentDaysCount > 0 
      ? (totalWithoutIncident / nonIncidentDaysCount).toFixed(1)
      : '0.0';

    const critical = list.filter(t => t.priority.toLowerCase() === 'critical').length;
    const high = list.filter(t => t.priority.toLowerCase() === 'high').length;
    const resolved = list.filter(t => {
      const status = t.ticket_status.toLowerCase();
      return status === 'closed' || status === 'resolved' || t.ResolvedAt || t.CompleteTime;
    }).length;

    const resolvedRate = total > 0 ? Math.round((resolved / total) * 100) : 0;

    // Compute average resolution time
    let durationSum = 0;
    let countedResolved = 0;
    list.forEach(t => {
      const endVal = t.ResolvedAt || t.CompleteTime || t.ClosedDate;
      if (endVal) {
        const start = new Date(t.CreatedDate).getTime();
        const end = new Date(endVal).getTime();
        if (start && end && end >= start) {
          durationSum += (end - start);
          countedResolved++;
        }
      }
    });

    const avgHrs = countedResolved > 0 
      ? (durationSum / countedResolved / (1000 * 60 * 60)).toFixed(1) 
      : '0.0';

    return { total, totalWithoutIncident, avgTicketsPerNonIncidentDay, critical, high, resolvedRate, avgHrs };
  });

  // Chart data calculations
  const priorityChartData = $derived.by(() => {
    const list = tickets;
    const result = { critical: 0, high: 0, medium: 0, low: 0 };
    list.forEach(t => {
      const p = t.priority.toLowerCase();
      if (p === 'critical') result.critical++;
      else if (p === 'high') result.high++;
      else if (p === 'medium') result.medium++;
      else if (p === 'low') result.low++;
    });
    return result;
  });


  const dailyTrendData = $derived.by(() => {
    // Group tickets in selected month by day
    const list = tickets;
    const days: Record<number, number> = {};
    
    list.forEach(t => {
      const date = new Date(t.CreatedDate);
      if (!isNaN(date.getTime())) {
        const day = date.getDate();
        days[day] = (days[day] || 0) + 1;
      }
    });

    // Get max day of current month (default to 31)
    const maxDay = 31;
    const dataPoints = [];
    for (let i = 1; i <= maxDay; i++) {
      dataPoints.push({ day: i, count: days[i] || 0 });
    }
    return dataPoints;
  });

  const maxDailyCount = $derived(
    Math.max(...dailyTrendData.map(d => d.count), 1)
  );

  // Dynamic lists of unique category values
  const uniqueT1 = $derived.by(() => {
    const set = new Set<string>();
    tickets.forEach(t => {
      if (t.ProductT1) set.add(t.ProductT1);
    });
    return ['All T1', ...Array.from(set).sort()];
  });

  const uniqueT2 = $derived.by(() => {
    const set = new Set<string>();
    tickets.forEach(t => {
      if (selectedT1 !== 'All T1' && t.ProductT1 !== selectedT1) return;
      if (t.ProductT2) set.add(t.ProductT2);
    });
    return ['All T2', ...Array.from(set).sort()];
  });

  const uniqueT3 = $derived.by(() => {
    const set = new Set<string>();
    tickets.forEach(t => {
      if (selectedT1 !== 'All T1' && t.ProductT1 !== selectedT1) return;
      if (selectedT2 !== 'All T2' && t.ProductT2 !== selectedT2) return;
      if (t.ProductT3) set.add(t.ProductT3);
    });
    return ['All T3', ...Array.from(set).sort()];
  });

  // Automatically reset dependent category selections when parent changes
  $effect(() => {
    if (selectedT1) {
      if (!uniqueT2.includes(selectedT2)) {
        selectedT2 = 'All T2';
      }
    }
  });

  $effect(() => {
    if (selectedT2) {
      if (!uniqueT3.includes(selectedT3)) {
        selectedT3 = 'All T3';
      }
    }
  });

  // Drag and drop handlers
  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
      validateAndSetFile(e.dataTransfer.files[0]);
    }
  }

  function handleFileSelect(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      validateAndSetFile(target.files[0]);
    }
  }

  function validateAndSetFile(file: File) {
    const ext = file.name.split('.').pop()?.toLowerCase();
    if (ext !== 'xlsx') {
      uploadError = 'Invalid file type. Only Excel spreadsheets (.xlsx) are supported.';
      uploadState = 'error';
      return;
    }
    selectedFile = file;
    uploadState = 'selected';
    uploadError = '';
  }

  // Upload actions
  async function triggerUpload() {
    if (!selectedFile) return;
    uploadState = 'uploading';
    
    const formData = new FormData();
    formData.append('file', selectedFile);

    try {
      const res = await fetch(`${API_URL}/upload`, {
        method: 'POST',
        body: formData,
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Server rejected the file upload.');
      }

      const data = await res.json();
      uploadStats = {
        inserted: data.inserted,
        skipped: data.skipped,
        filename: data.file
      };
      uploadState = 'success';
      
      // Auto refresh data and hide modal / upload state after delay
      setTimeout(async () => {
        await initializeData();
        closeUploadView();
      }, 2000);

    } catch (e: any) {
      uploadError = e.message || 'An error occurred during file transfer.';
      uploadState = 'error';
    }
  }

  function resetUpload() {
    selectedFile = null;
    uploadState = 'idle';
    uploadError = '';
  }

  function closeUploadView() {
    resetUpload();
    showUploadModal = false;
  }

  // Date formatter helper
  function formatDate(dateStr: string) {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  // Sort toggle helper
  function toggleSort(field: keyof Ticket) {
    if (sortField === field) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      sortField = field;
      sortDirection = 'desc';
    }
  }
</script>

<svelte:window onclick={(e) => {
  // @ts-ignore
  if (isMonthDropdownOpen && !e.target.closest('.month-dropdown-wrapper')) {
    isMonthDropdownOpen = false;
  }
}} />

<svelte:head>
  <title>IAM Ticket Analytics Hub</title>
  <meta name="description" content="State-of-the-art enterprise IAM ticket analysis and reporting platform." />
</svelte:head>

<!-- Background glow components -->
<div class="fixed inset-0 -z-50 bg-slate-950 overflow-hidden">
  <div class="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/10 rounded-full blur-[128px]"></div>
  <div class="absolute top-1/3 -right-40 w-[500px] h-[500px] bg-cyan-600/10 rounded-full blur-[160px]"></div>
  <div class="absolute -bottom-40 left-1/3 w-[600px] h-[600px] bg-emerald-600/5 rounded-full blur-[180px]"></div>
  <!-- Futuristic Grid -->
  <div class="absolute inset-0 bg-[linear-gradient(to_right,#0f172a_1px,transparent_1px),linear-gradient(to_bottom,#0f172a_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] opacity-30"></div>
</div>

{#if isLoading}
  <!-- Main Loader Screen -->
  <div class="flex flex-col items-center justify-center min-h-screen space-y-6" in:fade>
    <div class="relative w-20 h-20">
      <div class="absolute inset-0 border-4 border-indigo-500/20 rounded-full"></div>
      <div class="absolute inset-0 border-4 border-t-indigo-500 rounded-full animate-spin"></div>
      <div class="absolute inset-2 border-4 border-b-cyan-400 rounded-full animate-spin duration-700"></div>
    </div>
    <div class="text-center">
      <h2 class="text-xl font-medium text-slate-200">Loading system metrics</h2>
      <p class="text-xs text-slate-500 mt-1.5 font-mono">Initializing connection to tickets.db...</p>
    </div>
  </div>

{:else if tickets.length === 0 && !showUploadModal}
  <!-- Landing / Empty State File Upload Page -->
  <main class="min-h-screen flex flex-col items-center justify-center px-4 py-16" in:fade>
    <div class="w-full max-w-3xl space-y-12">
      
      <!-- Top Brand Header -->
      <div class="text-center space-y-4">
        <div class="inline-flex items-center gap-2.5 px-4 py-1.5 rounded-full border border-indigo-500/30 bg-indigo-500/10 text-xs font-semibold uppercase tracking-[0.2em] text-indigo-300">
          <span class="w-2 h-2 rounded-full bg-indigo-400 animate-pulse"></span>
          IAM Operations Center
        </div>
        <h1 class="text-4xl sm:text-6xl font-extrabold tracking-tight text-white leading-none bg-gradient-to-r from-white via-slate-100 to-slate-400 bg-clip-text">
          Ticket Analytics Hub
        </h1>
        <p class="text-base sm:text-lg text-slate-400 max-w-xl mx-auto font-light leading-relaxed">
          Unlock high-fidelity insights, trends, and performance metrics from your Identity & Access Management ticket data.
        </p>
      </div>

      <!-- Gorgeous Drag & Drop Glassmorphic Card -->
      <div 
        class="relative group rounded-3xl border transition-all duration-500 backdrop-blur-xl p-1.5 overflow-hidden
          {isDragging 
            ? 'border-indigo-400 bg-indigo-950/20 shadow-2xl shadow-indigo-500/10 scale-[1.01]' 
            : 'border-slate-800 bg-slate-900/30 hover:border-slate-700/80 shadow-2xl shadow-black/40'}"
        ondragover={handleDragOver}
        ondragleave={handleDragLeave}
        ondrop={handleDrop}
        role="button"
        tabindex="0"
      >
        <!-- Decorative Glow -->
        <div class="absolute -inset-1 px-10 py-10 bg-gradient-to-r from-indigo-500/10 via-cyan-500/5 to-emerald-500/10 blur-xl opacity-50 group-hover:opacity-100 transition duration-500 -z-10"></div>
        
        <div class="rounded-[22px] border border-slate-900/40 bg-slate-950/40 p-10 sm:p-14 text-center space-y-8 flex flex-col items-center">
          
          <!-- Cloud Upload Graphic -->
          <div class="relative flex items-center justify-center w-20 h-20 rounded-2xl bg-slate-900 border border-slate-800 group-hover:border-indigo-500/40 transition duration-500">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-10 h-10 text-indigo-400 group-hover:text-cyan-400 transition duration-500">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
            </svg>
            <div class="absolute -bottom-1 -right-1 w-5 h-5 rounded-full bg-emerald-500/20 border border-emerald-500 flex items-center justify-center animate-pulse">
              <span class="w-2.5 h-2.5 rounded-full bg-emerald-400"></span>
            </div>
          </div>

          <!-- Main Call to Actions -->
          <div class="space-y-3">
            {#if uploadState === 'idle'}
              <h3 class="text-xl font-semibold text-slate-200">Initialize Dashboard</h3>
              <p class="text-sm text-slate-400 max-w-sm mx-auto font-light">
                Drag and drop your spreadsheet file here, or click below to select from your device.
              </p>
            {:else if uploadState === 'selected' && selectedFile}
              <div class="space-y-2" in:slide>
                <h3 class="text-xl font-semibold text-emerald-400 flex items-center justify-center gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6">
                    <path fill-rule="evenodd" d="M2.25 12c0-5.385 4.365-9.75 9.75-9.75s9.75 4.365 9.75 9.75-4.365 9.75-9.75 9.75S2.25 17.385 2.25 12zm13.36-1.814a.75.75 0 10-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.74-5.249z" clip-rule="evenodd" />
                  </svg>
                  File Ready for Import
                </h3>
                <p class="text-sm font-mono text-slate-200 font-semibold bg-slate-900 border border-slate-800 px-4 py-2 rounded-xl inline-block shadow-inner max-w-md truncate">
                  {selectedFile.name}
                </p>
                <p class="text-xs text-slate-500">
                  Size: {(selectedFile.size / 1024).toFixed(1)} KB
                </p>
              </div>
            {:else if uploadState === 'uploading'}
              <div class="space-y-4" in:slide>
                <div class="flex items-center justify-center space-x-2">
                  <span class="w-3 h-3 bg-indigo-400 rounded-full animate-bounce"></span>
                  <span class="w-3 h-3 bg-cyan-400 rounded-full animate-bounce [animation-delay:0.2s]"></span>
                  <span class="w-3 h-3 bg-emerald-400 rounded-full animate-bounce [animation-delay:0.4s]"></span>
                </div>
                <h3 class="text-lg font-semibold text-slate-200">Importing records...</h3>
                <p class="text-xs text-slate-500 font-mono">Parsing sheets, running transactions, creating DB indexes</p>
              </div>
            {:else if uploadState === 'success'}
              <div class="space-y-3" in:slide>
                <div class="w-12 h-12 rounded-full bg-emerald-500/20 border border-emerald-500 flex items-center justify-center mx-auto text-emerald-400">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-6 h-6">
                    <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
                  </svg>
                </div>
                <h3 class="text-xl font-bold text-emerald-400">Success!</h3>
                <p class="text-sm text-slate-300 font-medium">
                  Parsed & imported <span class="text-emerald-400 font-semibold">{uploadStats.inserted}</span> tickets.
                </p>
                {#if uploadStats.skipped > 0}
                  <p class="text-xs text-amber-400">Skipped {uploadStats.skipped} invalid rows.</p>
                {/if}
              </div>
            {:else if uploadState === 'error'}
              <div class="space-y-3" in:slide>
                <div class="w-12 h-12 rounded-full bg-rose-500/20 border border-rose-500 flex items-center justify-center mx-auto text-rose-400 animate-shake">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-6 h-6">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
                  </svg>
                </div>
                <h3 class="text-lg font-bold text-rose-400">Parsing Failed</h3>
                <p class="text-xs text-rose-300 font-mono bg-rose-950/30 border border-rose-900/50 p-3 rounded-xl max-w-md mx-auto whitespace-pre-wrap">
                  {uploadError}
                </p>
              </div>
            {/if}
          </div>

          <!-- Buttons Section -->
          <div class="flex flex-col sm:flex-row items-center gap-4 w-full max-w-sm justify-center pt-2">
            {#if uploadState === 'idle'}
              <label class="cursor-pointer inline-flex items-center justify-center px-6 py-3 rounded-2xl bg-indigo-600 hover:bg-indigo-500 text-white font-medium shadow-lg hover:shadow-indigo-600/30 transition-all duration-300 w-full sm:w-auto text-sm border border-indigo-400/20">
                Browse Files
                <input type="file" accept=".xlsx" class="hidden" onchange={handleFileSelect} />
              </label>
            {:else if uploadState === 'selected'}
              <button 
                onclick={triggerUpload}
                class="px-6 py-3 rounded-2xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium shadow-lg hover:shadow-emerald-600/30 transition-all duration-300 w-full sm:w-auto text-sm border border-emerald-400/20"
              >
                Import Data
              </button>
              <button 
                onclick={resetUpload}
                class="px-6 py-3 rounded-2xl bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition-all duration-300 w-full sm:w-auto text-sm"
              >
                Cancel
              </button>
            {:else if uploadState === 'error'}
              <button 
                onclick={resetUpload}
                class="px-6 py-3 rounded-2xl bg-indigo-600 hover:bg-indigo-500 text-white font-medium transition-all duration-300 w-full sm:w-auto text-sm"
              >
                Retry
              </button>
            {/if}
          </div>

          <!-- Technical Spec List -->
          {#if uploadState === 'idle'}
            <div class="border-t border-slate-900/80 w-full pt-8 grid grid-cols-2 gap-4 text-left max-w-lg">
              <div>
                <p class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Required Columns</p>
                <p class="text-[11px] text-slate-500 mt-1 font-mono leading-relaxed">Ticket ID, Priority, Created Date, Ticket Type</p>
              </div>
              <div>
                <p class="text-xs font-semibold text-slate-400 uppercase tracking-wider">File Format</p>
                <p class="text-[11px] text-slate-500 mt-1 font-mono leading-relaxed">Microsoft Excel (.xlsx)</p>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </main>

{:else}
  <!-- Main Dashboard View -->
  <div class="min-h-screen flex flex-col" in:fade>
    
    <!-- Top Modern Navigation/Header -->
    <header class="border-b border-slate-900/60 bg-slate-950/80 backdrop-blur-xl sticky top-0 z-30 px-6 py-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <!-- Dashboard Logo Graphic -->
        <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-cyan-500 flex items-center justify-center text-white shadow-md shadow-indigo-600/20">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
          </svg>
        </div>
        <div>
          <h1 class="text-base font-bold text-white tracking-tight">IAM Analytics</h1>
          <p class="text-[10px] text-slate-500 font-mono leading-none">V1.0.0 // Dashboard</p>
        </div>
      </div>

      <nav class="hidden sm:flex items-center gap-1 bg-slate-900 border border-slate-850 p-1 rounded-xl text-xs font-semibold">
        <a href="/" class="px-3.5 py-2 rounded-lg bg-indigo-600 text-white shadow-md shadow-indigo-600/10 transition flex items-center gap-1.5">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
          </svg>
          <span>Home</span>
        </a>
        <a href="/analysis" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Analytics</a>
        <a href="/leaderboard" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Leaderboard</a>
        <a href="/incident" class="px-4 py-2 rounded-lg text-slate-400 hover:text-slate-200 transition">Incidents</a>
      </nav>

      <div class="flex items-center gap-3">
        <!-- Re-upload Data Trigger -->
        <button 
          onclick={() => showUploadModal = true}
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-200 border border-slate-800 hover:border-slate-700 font-medium text-xs tracking-wide transition-all duration-300"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-4 h-4 text-indigo-400">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
          </svg>
          Sync Data
        </button>
      </div>
    </header>

    <main class="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">

      <!-- Top Title & Action Row (Scrolls with page, aligned below header actions) -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-900/40 pb-4">
        <div>
          <h2 class="text-xl font-bold tracking-tight text-white flex items-center gap-2">Dashboard</h2>
          <p class="text-xs text-slate-400 font-light">Real-time indicators, priority splits, and daily created ticket volumes.</p>
        </div>

        {#if months.length > 0}
          <div class="relative month-dropdown-wrapper inline-block">
            <button 
              onclick={() => isMonthDropdownOpen = !isMonthDropdownOpen}
              class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 hover:border-slate-700/80 hover:bg-slate-850 text-slate-200 text-xs font-semibold focus:outline-none transition shadow"
            >
              <svg class="w-3.5 h-3.5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
              </svg>
              <span>{selectedMonth || 'Select Month'}</span>
              <svg class="w-3.5 h-3.5 text-slate-500 transition-transform duration-200 {isMonthDropdownOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>

            {#if isMonthDropdownOpen}
              <div 
                class="absolute right-0 mt-2 w-48 max-h-60 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-800 rounded-xl border border-slate-800 bg-slate-950/95 backdrop-blur-xl p-1.5 shadow-2xl z-40 space-y-0.5"
                transition:slide={{ duration: 150 }}
              >
                {#each months as month}
                  <button 
                    onclick={() => { selectedMonth = month; isMonthDropdownOpen = false; }}
                    class="w-full text-left px-3 py-2 rounded-lg text-xs font-medium transition flex items-center justify-between
                      {selectedMonth === month 
                        ? 'bg-indigo-600 text-white font-semibold shadow-sm' 
                        : 'text-slate-400 hover:bg-slate-900/60 hover:text-white'}"
                  >
                    <span>{month}</span>
                    {#if selectedMonth === month}
                      <svg class="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/></svg>
                    {/if}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Description Search Bar (Regex) -->
      <div class="rounded-2xl border border-slate-900 bg-slate-950/40 p-4">
        <div class="relative w-full">
          <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-4 h-4 text-slate-400">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.637 10.637z" />
            </svg>
          </span>
          <input
            type="text"
            bind:value={descriptionSearchInput}
            placeholder="Filter by description keyword or regex (e.g. ต่ออายุ, error.*failed, request.*access)..."
            class="w-full bg-slate-900 border border-slate-800 placeholder-slate-550 text-slate-200 rounded-xl pl-10 pr-4 py-2.5 text-xs focus:outline-none focus:border-indigo-500/80 focus:ring-1 focus:ring-indigo-500/20 shadow-inner"
          />
        </div>
      </div>

      <!-- Metrics Section -->
      <section class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        
        <!-- Total Tickets Card -->
        <div class="relative overflow-hidden rounded-2xl border border-slate-900 bg-slate-950 p-5 space-y-3">
          <div class="absolute -right-3 -top-3 w-16 h-16 bg-indigo-600/10 rounded-full blur-xl"></div>
          <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Total Volume</p>
          <div class="flex items-baseline space-x-2">
            <span class="text-3xl font-extrabold text-white tracking-tight">{metrics.total}</span>
            <span class="text-xs text-indigo-400 font-medium">tickets</span>
          </div>
          
          <!-- Baseline vs Incident Split Progress Bar -->
          <div class="h-1 bg-slate-900 rounded-full overflow-hidden relative">
            <div 
              class="h-full bg-indigo-500 rounded-full" 
              style="width: {metrics.total ? (metrics.totalWithoutIncident / metrics.total) * 100 : 0}%"
            ></div>
          </div>

          <!-- Secondary stats details -->
          <div class="space-y-1 text-[9px] font-mono pt-1.5 text-slate-500 border-t border-slate-900/60 mt-1">
            <div class="flex items-center justify-between">
              <span>Excl. Incidents:</span>
              <span class="font-bold text-indigo-300">{metrics.totalWithoutIncident}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>Avg/Day (Excl. Inc):</span>
              <span class="font-bold text-indigo-300">{metrics.avgTicketsPerNonIncidentDay}</span>
            </div>
          </div>
        </div>

        <!-- High/Critical Priority Card -->
        <div class="relative overflow-hidden rounded-2xl border border-slate-900 bg-slate-950 p-5 space-y-3">
          <div class="absolute -right-3 -top-3 w-16 h-16 bg-rose-600/10 rounded-full blur-xl"></div>
          <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Urgent & Critical</p>
          <div class="flex items-baseline space-x-2">
            <span class="text-3xl font-extrabold text-rose-400 tracking-tight">
              {metrics.critical + metrics.high}
            </span>
            <span class="text-[11px] text-slate-500">
              ({metrics.critical} crit / {metrics.high} high)
            </span>
          </div>
          <div class="h-1 bg-slate-900 rounded-full overflow-hidden">
            <div class="h-full bg-rose-500 rounded-full" style="width: {metrics.total ? Math.round(((metrics.critical + metrics.high) / metrics.total) * 100) : 0}%"></div>
          </div>
        </div>

        <!-- Resolution Rate Card -->
        <div class="relative overflow-hidden rounded-2xl border border-slate-900 bg-slate-950 p-5 space-y-3">
          <div class="absolute -right-3 -top-3 w-16 h-16 bg-emerald-600/10 rounded-full blur-xl"></div>
          <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Resolution Rate</p>
          <div class="flex items-baseline space-x-2">
            <span class="text-3xl font-extrabold text-emerald-400 tracking-tight">{metrics.resolvedRate}%</span>
            <span class="text-[11px] text-slate-500">closed or resolved</span>
          </div>
          <div class="h-1 bg-slate-900 rounded-full overflow-hidden">
            <div class="h-full bg-emerald-500 rounded-full" style="width: {metrics.resolvedRate}%"></div>
          </div>
        </div>

        <!-- Average Resolution Time Card -->
        <div class="relative overflow-hidden rounded-2xl border border-slate-900 bg-slate-950 p-5 space-y-3">
          <div class="absolute -right-3 -top-3 w-16 h-16 bg-cyan-600/10 rounded-full blur-xl"></div>
          <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Avg Resolution</p>
          <div class="flex items-baseline space-x-2">
            <span class="text-3xl font-extrabold text-cyan-400 tracking-tight">{metrics.avgHrs}</span>
            <span class="text-xs text-slate-500 font-medium">hours</span>
          </div>
          <div class="h-1 bg-slate-900 rounded-full overflow-hidden">
            <div class="h-full bg-cyan-400 rounded-full w-2/3"></div>
          </div>
        </div>

      </section>

      <!-- Charts & Visual Analytics Section -->
      <section class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        <!-- Priority Chart Card -->
        <div class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-5 space-y-6 flex flex-col justify-between">
          <div>
            <h3 class="text-sm font-bold text-slate-200">Volume by Priority</h3>
            <p class="text-xs text-slate-550">Critical items require immediate mitigation.</p>
          </div>
          
          <div class="space-y-4 py-2">
            {#each Object.entries(priorityChartData) as [prio, count]}
              <div class="space-y-1.5">
                <div class="flex items-center justify-between text-xs">
                  <span class="capitalize font-semibold text-slate-400 flex items-center gap-1.5">
                    {#if prio === 'critical'}
                      <span class="w-1.5 h-1.5 rounded-full bg-rose-500"></span>
                    {:else if prio === 'high'}
                      <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                    {:else if prio === 'medium'}
                      <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                    {:else}
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                    {/if}
                    {prio}
                  </span>
                  <span class="font-mono font-bold text-slate-300">{count}</span>
                </div>
                <div class="h-2 bg-slate-900 rounded-full overflow-hidden">
                  <div 
                    class="h-full rounded-full transition-all duration-1000 
                      {prio === 'critical' ? 'bg-rose-500' : ''} 
                      {prio === 'high' ? 'bg-amber-400' : ''} 
                      {prio === 'medium' ? 'bg-blue-500' : ''} 
                      {prio === 'low' ? 'bg-emerald-500' : ''}"
                    style="width: {metrics.total ? (count / metrics.total) * 100 : 0}%"
                  ></div>
                </div>
              </div>
            {/each}
          </div>
          <div class="text-[10px] text-slate-555 text-right leading-none">
            Updated via GORM
          </div>
        </div>

        <!-- Volume Daily Trend Chart Card -->
        <div class="lg:col-span-2 rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-5 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-sm font-bold text-slate-200">Daily Created Trend</h3>
              <p class="text-xs text-slate-550">
                Distribution of tickets generated across the days of the month.
                <span class="text-rose-400 font-semibold inline-flex items-center gap-1 ml-1">
                  <span class="w-1.5 h-1.5 rounded-full bg-rose-500 inline-block animate-pulse"></span>
                  Days with ≥50 tickets marked as Incident Days.
                </span>
              </p>
            </div>
          </div>

          <!-- Trend Bar Graph using Pure CSS & SVG -->
          <div class="h-44 flex items-end gap-1.5 pt-4 overflow-x-auto scrollbar-thin scrollbar-thumb-slate-800">
            {#each dailyTrendData as { day, count }}
              <div class="flex-1 flex flex-col items-center h-full justify-end group min-w-[20px]">
                <div class="relative w-full flex justify-center">
                  <!-- Hover tooltip -->
                  <div class="absolute bottom-full mb-1.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-slate-950/95 border border-slate-800 text-[10px] font-mono px-2 py-1 rounded-lg text-slate-200 z-10 whitespace-nowrap shadow-xl
                    {day <= 2 ? 'left-0 origin-left' : day >= 30 ? 'right-0 origin-right' : 'left-1/2 -translate-x-1/2'}">
                    {#if count >= 50}
                      <span class="text-rose-400 font-bold">⚠️ Incident Day:</span> {count} tickets
                    {:else}
                      {count} tickets
                    {/if}
                  </div>
                  
                  <!-- Blinking warning dot above the bar for Incident Days (>= 50 tickets) -->
                  {#if count >= 50}
                    <div class="absolute -top-3 flex items-center justify-center h-2 w-2">
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-rose-400 opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-rose-500"></span>
                    </div>
                  {/if}

                  <div 
                    class="w-full rounded-t-sm transition-all duration-500 group-hover:brightness-125
                      {count >= 50 
                        ? 'bg-gradient-to-t from-rose-600/90 to-rose-400 shadow-[0_0_10px_rgba(244,63,94,0.15)]' 
                        : count > 0 
                          ? 'bg-gradient-to-t from-indigo-600/80 to-indigo-400' 
                          : 'bg-slate-900/40'}"
                    style="height: {(count / maxDailyCount) * 120}px"
                  ></div>
                </div>
                <span class="text-[9px] font-mono text-slate-500 mt-2">{day}</span>
              </div>
            {/each}
          </div>
      </section>

      <!-- Search, Filters, and Ticket Database Section -->
      <section class="rounded-2xl border border-slate-900 bg-slate-950/40 backdrop-blur-xl p-5 space-y-5">
        
        <div class="flex flex-col md:flex-row gap-4 items-center justify-between">
          <div>
            <h3 class="text-sm font-bold text-slate-200">Ticket Register</h3>
            <p class="text-xs text-slate-500">Live search & filter your active database registers.</p>
          </div>

          <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
            
            <!-- Custom Search Box -->
            <div class="relative flex-1 md:flex-initial md:w-64">
              <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-4 h-4">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.637 10.637z" />
                </svg>
              </span>
              <input 
                type="text" 
                bind:value={searchQuery}
                placeholder="Search ticket, assignee, customer..." 
                class="w-full bg-slate-900 border border-slate-800 placeholder-slate-500 text-slate-200 rounded-xl pl-9 pr-4 py-2 text-xs focus:outline-none focus:border-indigo-500/80 focus:ring-1 focus:ring-indigo-500/20"
              />
            </div>

            <!-- Priority filters pill -->
            <div class="flex bg-slate-900 border border-slate-800 p-0.5 rounded-xl text-xs font-medium">
              {#each ['All', 'Critical', 'High', 'Medium', 'Low'] as prio}
                <button 
                  onclick={() => { selectedPriority = prio; currentPage = 1; }}
                  class="px-3 py-1.5 rounded-lg transition duration-200
                    {selectedPriority === prio 
                      ? 'bg-indigo-600 text-white shadow-sm shadow-indigo-600/10' 
                      : 'text-slate-400 hover:text-slate-200'}"
                >
                  {prio}
                </button>
              {/each}
            </div>

          </div>
        </div>

        <!-- Categories T1, T2, T3 Dropdowns -->
        <div class="flex flex-wrap items-center gap-3 w-full border-t border-slate-900/60 pt-4">
          <span class="text-xs text-slate-500 font-semibold uppercase tracking-wider">Product Categories:</span>
          
          <!-- Category T1 -->
          <div class="relative w-full sm:w-52">
            <select 
              bind:value={selectedT1}
              onchange={() => { currentPage = 1; }}
              class="w-full bg-slate-900 border border-slate-800 text-slate-200 rounded-xl px-3.5 py-2 text-xs focus:outline-none focus:border-indigo-500 cursor-pointer appearance-none"
            >
              {#each uniqueT1 as cat}
                <option value={cat}>{cat}</option>
              {/each}
            </select>
            <span class="absolute inset-y-0 right-0 flex items-center pr-3.5 pointer-events-none text-slate-500">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M19 9l-7 7-7-7"/></svg>
            </span>
          </div>

          <!-- Category T2 -->
          <div class="relative w-full sm:w-52">
            <select 
              bind:value={selectedT2}
              disabled={selectedT1 === 'All T1'}
              onchange={() => { currentPage = 1; }}
              class="w-full bg-slate-900 border border-slate-800 text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed rounded-xl px-3.5 py-2 text-xs focus:outline-none focus:border-indigo-500 cursor-pointer appearance-none"
            >
              {#each uniqueT2 as cat}
                <option value={cat}>{cat}</option>
              {/each}
            </select>
            <span class="absolute inset-y-0 right-0 flex items-center pr-3.5 pointer-events-none text-slate-500">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M19 9l-7 7-7-7"/></svg>
            </span>
          </div>

          <!-- Category T3 -->
          <div class="relative w-full sm:w-52">
            <select 
              bind:value={selectedT3}
              disabled={selectedT2 === 'All T2' || selectedT1 === 'All T1'}
              onchange={() => { currentPage = 1; }}
              class="w-full bg-slate-900 border border-slate-800 text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed rounded-xl px-3.5 py-2 text-xs focus:outline-none focus:border-indigo-500 cursor-pointer appearance-none"
            >
              {#each uniqueT3 as cat}
                <option value={cat}>{cat}</option>
              {/each}
            </select>
            <span class="absolute inset-y-0 right-0 flex items-center pr-3.5 pointer-events-none text-slate-500">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M19 9l-7 7-7-7"/></svg>
            </span>
          </div>

          {#if selectedT1 !== 'All T1' || selectedT2 !== 'All T2' || selectedT3 !== 'All T3'}
            <button 
              onclick={() => { selectedT1 = 'All T1'; selectedT2 = 'All T2'; selectedT3 = 'All T3'; currentPage = 1; }}
              class="text-xs text-indigo-400 hover:text-indigo-300 font-semibold transition ml-auto flex items-center gap-1.5"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
                <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.254-.674A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
              </svg>
              Clear Categories
            </button>
          {/if}
        </div>

        <!-- Desktop/Responsive Ticket Grid Table -->
        <div class="overflow-x-auto rounded-xl border border-slate-900/60 bg-slate-950/20">
          <table class="w-full border-collapse text-left text-xs">
            <thead>
              <tr class="border-b border-slate-900 bg-slate-950 text-slate-400 font-semibold">
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('ticket_id')}>
                  Ticket ID
                  {#if sortField === 'ticket_id'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('CreatedDate')}>
                  Created Date
                  {#if sortField === 'CreatedDate'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('subject')}>
                  Subject
                  {#if sortField === 'subject'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('priority')}>
                  Priority
                  {#if sortField === 'priority'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('ticket_status')}>
                  Status
                  {#if sortField === 'ticket_status'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
                <th class="p-4 cursor-pointer hover:bg-slate-900 transition" onclick={() => toggleSort('Assignee')}>
                  Assignee
                  {#if sortField === 'Assignee'}{sortDirection === 'asc' ? ' ↑' : ' ↓'}{/if}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-900">
              {#if paginatedTickets.length === 0}
                <tr>
                  <td colspan="6" class="p-8 text-center text-slate-500 font-mono">No matching records found.</td>
                </tr>
              {:else}
                {#each paginatedTickets as ticket}
                  <tr 
                    onclick={() => selectedTicket = ticket}
                    class="hover:bg-slate-900/30 transition cursor-pointer border-b border-slate-900/30"
                  >
                    <td class="p-4 font-mono font-bold text-slate-300">{ticket.ticket_id}</td>
                    <td class="p-4 text-slate-400">{formatDate(ticket.CreatedDate)}</td>
                    <td class="p-4 font-medium text-slate-200 max-w-xs truncate">{ticket.subject}</td>
                    <td class="p-4">
                      <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold border capitalize
                        {ticket.priority.toLowerCase() === 'critical' ? 'bg-rose-500/10 border-rose-500/30 text-rose-400' : ''}
                        {ticket.priority.toLowerCase() === 'high' ? 'bg-amber-500/10 border-amber-500/30 text-amber-400' : ''}
                        {ticket.priority.toLowerCase() === 'medium' ? 'bg-blue-500/10 border-blue-500/30 text-blue-400' : ''}
                        {ticket.priority.toLowerCase() === 'low' ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : ''}"
                      >
                        {ticket.priority}
                      </span>
                    </td>
                    <td class="p-4">
                      <span class="inline-flex items-center gap-1.5 text-slate-300">
                        <span class="w-1.5 h-1.5 rounded-full 
                          {ticket.ticket_status.toLowerCase() === 'closed' || ticket.ticket_status.toLowerCase() === 'resolved' 
                            ? 'bg-emerald-400' 
                            : 'bg-amber-400 animate-pulse'}"
                        ></span>
                        {ticket.ticket_status}
                      </span>
                    </td>
                    <td class="p-4 text-slate-400">{ticket.Assignee || '-'}</td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>

        <!-- Pagination Controls -->
        {#if totalPages > 1}
          <div class="flex items-center justify-between border-t border-slate-900/50 pt-4">
            <span class="text-xs text-slate-500">
              Showing page <span class="text-slate-300 font-semibold">{currentPage}</span> of <span class="text-slate-300 font-semibold">{totalPages}</span>
            </span>
            <div class="flex gap-2">
              <button 
                onclick={() => currentPage = Math.max(currentPage - 1, 1)} 
                disabled={currentPage === 1}
                class="px-3.5 py-1.5 rounded-lg border border-slate-800 bg-slate-900 text-slate-300 disabled:opacity-30 disabled:pointer-events-none hover:bg-slate-850 text-xs transition duration-200"
              >
                Previous
              </button>
              <button 
                onclick={() => currentPage = Math.min(currentPage + 1, totalPages)} 
                disabled={currentPage === totalPages}
                class="px-3.5 py-1.5 rounded-lg border border-slate-800 bg-slate-900 text-slate-300 disabled:opacity-30 disabled:pointer-events-none hover:bg-slate-850 text-xs transition duration-200"
              >
                Next
              </button>
            </div>
          </div>
        {/if}

      </section>

    </main>

    <!-- Footer -->
    <footer class="border-t border-slate-900/40 bg-slate-950 py-6 px-6 text-center text-[10px] text-slate-600 font-mono">
      IAM Ticket Dashboard // Backend SQLite 3 & Gin API // Developed Pair-wise
    </footer>

  </div>
{/if}

<!-- Detail Drawer Panel -->
{#if selectedTicket}
  <!-- Overlay -->
  <div 
    onclick={() => selectedTicket = null}
    class="fixed inset-0 bg-slate-950/70 backdrop-blur-sm z-40" 
    transition:fade
    role="button"
    tabindex="0"
  ></div>

  <!-- Drawer Body -->
  <div 
    class="fixed right-0 top-0 bottom-0 w-full max-w-lg bg-slate-950 border-l border-slate-900 z-50 p-6 flex flex-col justify-between shadow-2xl overflow-y-auto"
    transition:fly={{ x: 300, duration: 300 }}
  >
    <div class="space-y-6">
      
      <!-- Drawer Header -->
      <div class="flex items-center justify-between border-b border-slate-900/80 pb-4">
        <div>
          <span class="text-[10px] font-mono bg-indigo-500/10 border border-indigo-500/30 text-indigo-400 px-2 py-0.5 rounded font-bold">
            {selectedTicket.ticket_id}
          </span>
          <h2 class="text-sm font-bold text-slate-200 mt-2">{selectedTicket.subject}</h2>
        </div>
        <button 
          onclick={() => selectedTicket = null}
          class="w-8 h-8 rounded-lg bg-slate-900 hover:bg-slate-800 text-slate-400 flex items-center justify-center border border-slate-800"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
            <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
          </svg>
        </button>
      </div>

      <!-- Drawer Content Grid -->
      <div class="space-y-4 text-xs">
        
        <!-- Section: Description -->
        <div class="space-y-1">
          <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Detailed Description</h4>
          <p class="bg-slate-900/40 border border-slate-900/60 rounded-xl p-3.5 text-slate-300 leading-relaxed font-light whitespace-pre-line">
            {selectedTicket.DetailDescription || 'No description provided.'}
          </p>
        </div>

        <!-- Grid Attributes -->
        <div class="grid grid-cols-2 gap-4 pt-2">
          
          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Ticket Type</h4>
            <p class="text-slate-250 font-medium">{selectedTicket.ticket_type}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Related Type</h4>
            <p class="text-slate-250 font-medium capitalize">{selectedTicket.related_type || 'None (Normal Request)'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Priority</h4>
            <p class="text-slate-250 font-medium capitalize">{selectedTicket.priority}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Status</h4>
            <p class="text-slate-250 font-medium capitalize">{selectedTicket.ticket_status}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Customer Name</h4>
            <p class="text-slate-250 font-medium">{selectedTicket.customer_name || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Assignee</h4>
            <p class="text-slate-250 font-medium">{selectedTicket.Assignee || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Assignment Group</h4>
            <p class="text-slate-250 font-medium">{selectedTicket.GroupAssignee || '-'}</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Company / Country</h4>
            <p class="text-slate-250 font-medium">{selectedTicket.company} ({selectedTicket.country})</p>
          </div>

          <div class="space-y-1">
            <h4 class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Site / Region</h4>
            <p class="text-slate-250 font-medium">
              {selectedTicket.site || '-'} / {selectedTicket.region_site || '-'}
            </p>
          </div>

        </div>

        <!-- Categorization -->
        <div class="border-t border-slate-900/60 pt-4 grid grid-cols-3 gap-2">
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T1</h4>
            <p class="text-slate-300 font-medium text-[11px] truncate" title={selectedTicket.ProductT1}>{selectedTicket.ProductT1 || '-'}</p>
          </div>
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T2</h4>
            <p class="text-slate-300 font-medium text-[11px] truncate" title={selectedTicket.ProductT2}>{selectedTicket.ProductT2 || '-'}</p>
          </div>
          <div class="space-y-1">
            <h4 class="text-[10px] font-semibold text-slate-500 uppercase tracking-wider">Category T3</h4>
            <p class="text-slate-300 font-medium text-[11px] truncate" title={selectedTicket.ProductT3}>{selectedTicket.ProductT3 || '-'}</p>
          </div>
        </div>

        <!-- Timestamps -->
        <div class="border-t border-slate-900/60 pt-4 space-y-2">
          <div class="flex justify-between items-center text-[11px]">
            <span class="text-slate-500">Created At</span>
            <span class="font-mono text-slate-300">{formatDate(selectedTicket.CreatedDate)}</span>
          </div>
          {#if selectedTicket.ResolvedAt}
            <div class="flex justify-between items-center text-[11px]">
              <span class="text-slate-500">Resolved At</span>
              <span class="font-mono text-slate-300">{formatDate(selectedTicket.ResolvedAt)}</span>
            </div>
          {/if}
          {#if selectedTicket.CompleteTime}
            <div class="flex justify-between items-center text-[11px]">
              <span class="text-slate-500">Completed At</span>
              <span class="font-mono text-slate-300">{formatDate(selectedTicket.CompleteTime)}</span>
            </div>
          {/if}
          <div class="flex justify-between items-center text-[11px]">
            <span class="text-slate-500">Metadata Source</span>
            <span class="font-mono text-slate-400 bg-slate-900 border border-slate-800 px-2 py-0.5 rounded truncate max-w-xs">{selectedTicket.SourceFile}</span>
          </div>
        </div>

      </div>

    </div>

    <div class="pt-6 border-t border-slate-900/80 mt-6 flex justify-end">
      <button 
        onclick={() => selectedTicket = null}
        class="px-5 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition"
      >
        Close View
      </button>
    </div>
  </div>
{/if}

<!-- Re-upload Data Modal Overlay -->
{#if showUploadModal}
  <!-- Overlay -->
  <div 
    onclick={closeUploadView}
    class="fixed inset-0 bg-slate-950/80 backdrop-blur-md z-40" 
    transition:fade
    role="button"
    tabindex="0"
  ></div>

  <!-- Modal Dialog -->
  <div 
    class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-xl bg-slate-950 border border-slate-900 rounded-3xl z-50 p-6 sm:p-8 space-y-6 shadow-2xl shadow-black/80"
    transition:fly={{ y: 50, duration: 300 }}
  >
    <div class="flex items-center justify-between border-b border-slate-900/60 pb-3">
      <div>
        <h3 class="text-base font-bold text-white tracking-tight">Sync New Spreadsheet</h3>
        <p class="text-xs text-slate-500">This will overwrite the active ticket records in sqlite.db.</p>
      </div>
      <button 
        onclick={closeUploadView}
        class="w-8 h-8 rounded-lg bg-slate-900 hover:bg-slate-800 text-slate-400 flex items-center justify-center border border-slate-800"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
          <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
        </svg>
      </button>
    </div>

    <!-- Drag & Drop Uploader Component inside Modal -->
    <div 
      class="border rounded-2xl p-6 text-center space-y-6 transition duration-300
        {isDragging 
          ? 'border-indigo-400 bg-indigo-950/25' 
          : 'border-slate-800 bg-slate-900/10 hover:border-slate-700/80'}"
      ondragover={handleDragOver}
      ondragleave={handleDragLeave}
      ondrop={handleDrop}
      role="button"
      tabindex="0"
    >
      <div class="flex items-center justify-center w-14 h-14 rounded-xl bg-slate-900 border border-slate-800 mx-auto">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-6 h-6 text-indigo-400">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
        </svg>
      </div>

      <div class="space-y-2">
        {#if uploadState === 'idle'}
          <p class="text-xs text-slate-400 font-light">Drag & drop your Excel spreadsheet (.xlsx) or click to browse.</p>
        {:else if uploadState === 'selected' && selectedFile}
          <div class="space-y-1 bg-slate-900 border border-slate-800 p-2.5 rounded-xl max-w-sm mx-auto">
            <p class="text-xs font-mono font-bold text-slate-200 truncate">{selectedFile.name}</p>
            <p class="text-[10px] text-slate-500">{ (selectedFile.size / 1024).toFixed(1) } KB</p>
          </div>
        {:else if uploadState === 'uploading'}
          <p class="text-xs font-medium text-slate-200">Importing records in background...</p>
        {:else if uploadState === 'success'}
          <p class="text-xs font-bold text-emerald-400">Successfully synced {uploadStats.inserted} tickets!</p>
        {:else if uploadState === 'error'}
          <p class="text-xs text-rose-400 font-mono bg-rose-950/20 p-2.5 rounded-xl border border-rose-900/50 max-w-sm mx-auto truncate" title={uploadError}>
            {uploadError}
          </p>
        {/if}
      </div>

      <div class="flex items-center gap-3 justify-center">
        {#if uploadState === 'idle'}
          <label class="cursor-pointer inline-flex items-center px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs tracking-wide transition">
            Browse Files
            <input type="file" accept=".xlsx" class="hidden" onchange={handleFileSelect} />
          </label>
        {:else if uploadState === 'selected'}
          <button 
            onclick={triggerUpload}
            class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-xs tracking-wide transition"
          >
            Start Sync
          </button>
          <button 
            onclick={resetUpload}
            class="px-4 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-400 border border-slate-800 text-xs transition"
          >
            Cancel
          </button>
        {:else if uploadState === 'error'}
          <button 
            onclick={resetUpload}
            class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs transition"
          >
            Retry
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}