    document.addEventListener('DOMContentLoaded', function () {
        loadAttendanceClasses();

        // Search functionality
        let classSearchTimeout = null;
        const searchInput = document.getElementById('classSearchInput');
        if (searchInput) {
            searchInput.addEventListener('input', function () {
                const query = this.value.trim();
                if (classSearchTimeout) clearTimeout(classSearchTimeout);
                classSearchTimeout = setTimeout(() => loadAttendanceClasses(query), 300);
            });
        }

        // Set today's date as default if empty
        const dateInput = document.getElementById('attendanceDate');
        if (!dateInput.value) {
            const today = new Date().toISOString().split('T')[0];
            dateInput.value = today;
        }

        // Modal transition handling
        const modal = document.getElementById('classSelectionModal');
        const backdrop = document.getElementById('modalBackdrop');
        const panel = document.getElementById('modalPanel');

        window.openClassModal = function () {
            modal.classList.remove('hidden');
            // Small delay to allow display:block to apply before opacity transition
            setTimeout(() => {
                backdrop.classList.remove('opacity-0');
                panel.classList.remove('opacity-0', 'translate-y-4', 'sm:translate-y-0', 'sm:scale-95');
                panel.classList.add('opacity-100', 'translate-y-0', 'sm:scale-100');
            }, 10);
        };

        window.closeClassModal = function () {
            backdrop.classList.add('opacity-0');
            panel.classList.remove('opacity-100', 'translate-y-0', 'sm:scale-100');
            panel.classList.add('opacity-0', 'translate-y-4', 'sm:translate-y-0', 'sm:scale-95');

            setTimeout(() => {
                modal.classList.add('hidden');
            }, 300);
        };

        // Close modal on backdrop click
        modal.addEventListener('click', function (e) {
            if (e.target === modal || e.target === backdrop) {
                closeClassModal();
            }
        });

        // Close modal on escape key
        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape' && !modal.classList.contains('hidden')) {
                closeClassModal();
            }
        });
    });

    async function loadStudentCount(classId, element) {
        try {
            const response = await fetch('/api/attendance/class/' + classId);
            const data = await response.json();

            if (response.ok) {
                element.textContent = `${data.count} Students`;
                element.classList.add('text-indigo-600', 'bg-indigo-50', 'border-indigo-100');
                element.classList.remove('text-slate-400', 'bg-slate-50', 'border-slate-100');
            } else {
                element.textContent = '0 Students';
            }
        } catch (error) {
            console.error('Error loading student count:', error);
            element.textContent = 'Err';
        }
    }

    async function loadClassesForDate() {
        const dateInput = document.getElementById('attendanceDate');
        const selectedDate = dateInput.value;

        if (!selectedDate) {
            alert('Please select a date first');
            return;
        }

        // Show modal using new transition function
        openClassModal();
        document.getElementById('modalDateText').textContent = `Classes scheduled for ${selectedDate}`;

        // Reset list state
        const classesList = document.getElementById('modalClassesList');
        classesList.innerHTML = `
            <div class="flex flex-col items-center justify-center py-12">
                <div class="w-12 h-12 border-4 border-slate-100 border-t-indigo-600 rounded-full animate-spin mb-4"></div>
                <p class="text-xs font-bold text-slate-400 uppercase tracking-widest">Loading schedule...</p>
            </div>
        `;

        try {
            // Convert date to day of week
            const date = new Date(selectedDate);
            const dayOfWeek = date.toLocaleDateString('en-US', { weekday: 'long' }).toLowerCase();

            const container = document.getElementById('mainAttendanceContainer');
            const canAccessAllClasses = container ? container.dataset.userAccess === "true" : false;

            // Get lessons for the selected day (all or current user's)
            // Get lessons for the selected day (all or current user's)
            let endpoint = '/api/attendance/teacher-lessons/' + dayOfWeek;
            if (canAccessAllClasses) {
                endpoint = '/api/attendance/all-lessons/' + dayOfWeek;
            }
            const response = await fetch(endpoint);
            const data = await response.json();

            if (response.ok) {
                displayClassesInModal(data.timetable_entries || [], selectedDate);
            } else {
                console.error(data.error);
                displayModalError(data.error || 'Failed to load classes');
            }
        } catch (error) {
            console.error('Error loading classes:', error);
            displayModalError('Failed to load classes. Please try again.');
        }
    }

    function displayClassesInModal(lessons, date) {
        // Update lesson cache for lookup
        if (!window.lessonCache) window.lessonCache = {};
        lessons.forEach(l => window.lessonCache[l.id] = l);

        const classesList = document.getElementById('modalClassesList');
        classesList.innerHTML = '';

        if (lessons.length === 0) {
            classesList.innerHTML = `
            <div class="text-center py-12 bg-slate-50 rounded-2xl border border-dashed border-slate-200">
                <div class="w-14 h-14 bg-white rounded-2xl flex items-center justify-center text-slate-300 mx-auto mb-4 border border-slate-100 shadow-sm">
                    <i class="fas fa-calendar-times text-xl"></i>
                </div>
                <h4 class="text-sm font-black text-slate-800 uppercase tracking-tight mb-1">No Classes Scheduled</h4>
                <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">No classes found for ${date}</p>
            </div>
        `;
            return;
        }

        // Group lessons by class
        const classesByClass = {};
        lessons.forEach(lesson => {
            const classKey = lesson.class_id;
            if (!classesByClass[classKey]) {
                classesByClass[classKey] = [];
            }
            classesByClass[classKey].push(lesson);
        });

        // Display classes
        Object.keys(classesByClass).forEach(classId => {
            const classLessons = classesByClass[classId];
            const classCard = document.createElement('div');
            classCard.className = 'bg-white border border-slate-200 rounded-xl p-5 hover:border-indigo-200 hover:shadow-md transition-all group';

            const lessonsHtml = classLessons.map(lesson => `
            <div class="flex items-center justify-between py-3 px-4 bg-slate-50 rounded-lg mb-3 last:mb-0 border border-slate-100">
                <div class="flex items-center gap-4">
                    <div class="w-10 h-10 bg-white rounded-lg flex items-center justify-center border border-slate-200 text-indigo-600 shadow-sm font-bold text-[10px]">
                        ${lesson.time_slot.split(' - ')[0]}
                    </div>
                    <div>
                        <p class="text-[10px] font-black uppercase tracking-widest text-slate-500 mb-0.5">${lesson.time_slot}</p>
                        <p class="text-xs font-bold text-slate-800 flex items-center gap-2">
                            <span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>${lesson.subject_id}
                        </p>
                    </div>
                </div>
                <button onclick="takeAttendanceForLesson('${lesson.id}', '${date}')" 
                        class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-indigo-700 transition-all shadow-sm">
                    Take Roll
                </button>
            </div>
        `).join('');

            classCard.innerHTML = `
            <div class="flex items-center justify-between mb-4 pb-4 border-b border-slate-100">
                <div>
                    <h4 class="text-sm font-black text-slate-800 uppercase tracking-tight">Class: ${classId}</h4>
                    <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">${classLessons.length} lesson(s) scheduled</p>
                </div>
                <button onclick="takeClassAttendance('${classId}', '${date}')" 
                        class="px-3 py-1.5 bg-emerald-50 text-emerald-600 border border-emerald-100 rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-emerald-100 transition-all flex items-center gap-2">
                    <i class="fas fa-users"></i>
                    Full Class
                </button>
            </div>
            <div>
                ${lessonsHtml}
            </div>
        `;

            classesList.appendChild(classCard);
        });
    }

    function displayModalError(message) {
        const classesList = document.getElementById('modalClassesList');
        classesList.innerHTML = `
        <div class="text-center py-12 bg-rose-50 rounded-2xl border border-rose-100">
            <div class="w-14 h-14 bg-white rounded-2xl flex items-center justify-center text-rose-500 mx-auto mb-4 border border-rose-100 shadow-sm">
                <i class="fas fa-exclamation-triangle text-xl"></i>
            </div>
            <h4 class="text-sm font-black text-slate-800 uppercase tracking-tight mb-1">Unable to Load</h4>
            <p class="text-[10px] font-bold text-rose-500 uppercase tracking-widest">${message}</p>
        </div>
    `;
    }

    function takeClassAttendance(classId, date) {
        closeClassModal();
        // Small delay for animation
        setTimeout(() => {
            window.location.href = '/attendance/class/' + classId + '/date/' + date;
        }, 300);
    }

    function takeAttendanceForLesson(timetableEntryId, date) {
        const lesson = window.lessonCache[timetableEntryId];
        const params = new URLSearchParams({
            timetable_entry_id: timetableEntryId,
            date: date,
            lesson_info: JSON.stringify(lesson)
        });
        window.location.href = '/attendance/lesson?' + params.toString();
    }

    function loadTodayLessons() {
        const today = new Date().toISOString().split('T')[0];
        document.getElementById('attendanceDate').value = today;
        loadClassesForDate();
    }

    function loadUpcomingLessons() {
        window.location.href = '/attendance/timetable';
    }

    // Load classes table for attendance
    async function loadAttendanceClasses(search = '') {
        const tableBody = document.getElementById('attendanceTableBody');

        // Show loading state if it's not a fresh page load (optional, table starts with skeleton)
        if (search) {
            tableBody.innerHTML = `
                <tr class="skeleton-row">
                    <td colspan="5" class="px-5 py-12 text-center text-slate-400">
                        <div class="flex flex-col items-center justify-center space-y-3">
                            <i class="fas fa-spinner fa-spin text-2xl text-indigo-500 opacity-20"></i>
                            <span class="text-[10px] font-black uppercase tracking-widest opacity-50">Searching Nodes...</span>
                        </div>
                    </td>
                </tr>
            `;
        }

        try {
            let url = '/api/classes/table';
            if (search) {
                url = '/api/classes/table?search=' + encodeURIComponent(search);
            }
            const response = await fetch(url, { credentials: 'include' });

            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);

            const result = await response.json();

            if (result.success && result.classes) {
                displayAttendanceClasses(result.classes);
            } else {
                showAttendanceEmptyState();
            }
        } catch (error) {
            console.error('Error loading classes:', error);
            showAttendanceErrorState();
        }
    }

    function displayAttendanceClasses(classes) {
        const tableBody = document.getElementById('attendanceTableBody');

        if (classes.length === 0) {
            showAttendanceEmptyState();
            return;
        }

        tableBody.innerHTML = classes.map(classItem => `
            <tr class="hover:bg-indigo-50/30 transition-colors group border-b border-slate-50 last:border-b-0">
                <td class="px-5 py-4">
                    <div class="flex items-center">
                        <div class="w-10 h-10 rounded-xl bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-500 shadow-sm group-hover:scale-105 transition-transform mr-4">
                            <i class="fas fa-layer-group text-lg"></i>
                        </div>
                        <div>
                            <div class="text-[11px] font-black text-slate-700 uppercase tracking-wide group-hover:text-indigo-700 transition-colors">${classItem.name}</div>
                            <div class="text-[9px] font-bold text-slate-400 uppercase tracking-wider mt-0.5">
                                <span class="bg-slate-100 text-slate-500 px-1.5 py-0.5 rounded text-[8px] border border-slate-200">CODE: ${classItem.code || 'N/A'}</span>
                            </div>
                        </div>
                    </div>
                </td>
                <td class="px-5 py-4 text-left">
                    ${classItem.teacher ? `
                        <div class="flex items-center gap-2">
                            <div class="w-6 h-6 rounded-lg bg-indigo-100 flex items-center justify-center text-indigo-600 text-[10px] font-bold">
                                ${classItem.teacher.first_name.charAt(0)}${classItem.teacher.last_name.charAt(0)}
                            </div>
                            <div>
                                <div class="text-[10px] font-bold text-slate-700 uppercase tracking-wide">${classItem.teacher.first_name} ${classItem.teacher.last_name}</div>
                            </div>
                        </div>
                    ` : `
                        <span class="text-[9px] font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                            <i class="fas fa-user-slash"></i> Unassigned
                        </span>
                    `}
                </td>
                <td class="px-5 py-4 text-center">
                    ${classItem.student_count && classItem.student_count > 0 ? `
                        <div class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-indigo-50 border border-indigo-100">
                             <span class="text-[10px] font-black text-indigo-600">${classItem.student_count} Students</span>
                        </div>
                    ` : `
                        <span class="text-[9px] font-bold text-slate-300 uppercase tracking-wider">Empty Node</span>
                    `}
                </td>
                 <td class="px-5 py-4 text-center">
                   ${classItem.is_active ? `
                        <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-50 border border-emerald-100 text-emerald-600">
                            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                            <span class="text-[9px] font-black uppercase tracking-wider">Active</span>
                        </span>
                    ` : `
                        <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-50 border border-slate-200 text-slate-400">
                            <span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
                            <span class="text-[9px] font-black uppercase tracking-wider">Inactive</span>
                        </span>
                    `}
                </td>
                <td class="px-5 py-4 text-right">
                    <a href="/attendance/class/${classItem.id}" 
                       class="inline-flex items-center justify-center px-4 py-1.5 bg-white border border-slate-200 text-slate-600 rounded-lg text-[10px] font-black uppercase tracking-widest hover:bg-black hover:text-white hover:border-black transition-all shadow-sm">
                        Take Roll
                    </a>
                </td>
            </tr>
        `).join('');
    }

    function showAttendanceEmptyState() {
        document.getElementById('attendanceTableBody').innerHTML = `
            <tr>
                <td colspan="5" class="px-5 py-12 text-center text-slate-400">
                    <div class="flex flex-col items-center justify-center space-y-3">
                        <div class="w-16 h-16 bg-slate-50 rounded-2xl flex items-center justify-center border border-slate-100 mb-2">
                            <i class="fas fa-folder-open text-3xl text-slate-300"></i>
                         </div>
                        <h4 class="text-[11px] font-black text-slate-500 uppercase tracking-widest">Registry Empty</h4>
                        <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">No active classes found</p>
                    </div>
                </td>
            </tr>
        `;
    }

    function showAttendanceErrorState() {
        document.getElementById('attendanceTableBody').innerHTML = `
            <tr>
                <td colspan="5" class="px-5 py-12 text-center text-rose-400">
                    <div class="flex flex-col items-center justify-center space-y-2">
                        <i class="fas fa-exclamation-circle text-2xl mb-1"></i>
                        <span class="text-[10px] font-black uppercase tracking-widest">Connection Failure</span>
                    </div>
                </td>
            </tr>
        `;
    }
