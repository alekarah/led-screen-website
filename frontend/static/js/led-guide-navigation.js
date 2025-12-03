// LED Guide Navigation Features
// - Reading Progress Bar
// - Back to Top Button
// - Collapsible Advantages
// - Active ToC Link on Scroll

(function() {
    'use strict';

    // ===== Reading Progress Bar =====
    function initProgressBar() {
        const progressBar = document.querySelector('.reading-progress-bar');
        if (!progressBar) return;

        function updateProgressBar() {
            // Calculate scroll percentage
            const windowHeight = window.innerHeight;
            const documentHeight = document.documentElement.scrollHeight;
            const scrollTop = window.pageYOffset || document.documentElement.scrollTop;

            const scrollPercent = (scrollTop / (documentHeight - windowHeight)) * 100;
            progressBar.style.width = Math.min(scrollPercent, 100) + '%';
        }

        // Update on scroll
        window.addEventListener('scroll', updateProgressBar, { passive: true });

        // Initial update
        updateProgressBar();
    }

    // ===== Back to Top Button =====
    function initBackToTop() {
        const backToTopBtn = document.getElementById('backToTop');
        if (!backToTopBtn) return;

        function toggleBackToTop() {
            if (window.pageYOffset > 300) {
                backToTopBtn.classList.add('visible');
            } else {
                backToTopBtn.classList.remove('visible');
            }
        }

        // Show/hide on scroll
        window.addEventListener('scroll', toggleBackToTop, { passive: true });

        // Click handler
        backToTopBtn.addEventListener('click', function(e) {
            e.preventDefault();
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });

        // Initial check
        toggleBackToTop();
    }

    // ===== Collapsible Advantages =====
    function initCollapsibleAdvantages() {
        const toggleBtn = document.querySelector('.btn-toggle-advantages');
        if (!toggleBtn) return;

        const showText = toggleBtn.querySelector('.toggle-text-show');
        const hideText = toggleBtn.querySelector('.toggle-text-hide');

        let isExpanded = false;

        toggleBtn.addEventListener('click', function(e) {
            e.preventDefault();
            isExpanded = !isExpanded;

            const hiddenItems = document.querySelectorAll('.advantage-item--hidden');

            if (isExpanded) {
                hiddenItems.forEach(function(item) {
                    item.classList.remove('advantage-item--hidden');
                    item.classList.add('advantage-item--shown');
                });
                toggleBtn.classList.add('expanded');
                showText.style.display = 'none';
                hideText.style.display = 'inline';
            } else {
                const shownItems = document.querySelectorAll('.advantage-item--shown');
                shownItems.forEach(function(item) {
                    item.classList.add('advantage-item--hidden');
                    item.classList.remove('advantage-item--shown');
                });
                toggleBtn.classList.remove('expanded');
                showText.style.display = 'inline';
                hideText.style.display = 'none';
            }
        });
    }

    // ===== Active ToC Link on Scroll (Intersection Observer) =====
    function initActiveToCLinks() {
        const tocLinks = document.querySelectorAll('.toc-dropdown-link');
        if (!tocLinks || tocLinks.length === 0) return;

        // Create a map of section IDs to links
        const sections = [];
        tocLinks.forEach(function(link) {
            const href = link.getAttribute('href');
            if (href && href.startsWith('#')) {
                const sectionId = href.substring(1);
                const section = document.getElementById(sectionId);
                if (section) {
                    sections.push({
                        id: sectionId,
                        element: section,
                        link: link
                    });
                }
            }
        });

        if (sections.length === 0) return;

        // Intersection Observer to detect which section is in view
        const observerOptions = {
            root: null,
            rootMargin: '-20% 0px -70% 0px', // Trigger when section is 20% from top
            threshold: 0
        };

        const observer = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    // Remove active class from all links
                    tocLinks.forEach(function(link) {
                        link.classList.remove('active');
                    });

                    // Find the link for this section and mark it active
                    const activeSection = sections.find(function(s) {
                        return s.element === entry.target;
                    });

                    if (activeSection) {
                        activeSection.link.classList.add('active');
                    }
                }
            });
        }, observerOptions);

        // Observe all sections
        sections.forEach(function(section) {
            observer.observe(section.element);
        });
    }

    // ===== Smooth Scroll for ToC Links =====
    function initSmoothScrollToC() {
        const tocLinks = document.querySelectorAll('.toc-dropdown-link');

        tocLinks.forEach(function(link) {
            link.addEventListener('click', function(e) {
                const href = link.getAttribute('href');

                // Only handle hash links
                if (href && href.startsWith('#')) {
                    const targetId = href.substring(1);
                    const targetElement = document.getElementById(targetId);

                    if (targetElement) {
                        e.preventDefault();

                        // Smooth scroll to target
                        targetElement.scrollIntoView({
                            behavior: 'smooth',
                            block: 'start'
                        });

                        // Update URL hash without jump
                        if (history.pushState) {
                            history.pushState(null, null, href);
                        }
                    }
                }
            });
        });
    }

    // ===== Initialize All Features =====
    function init() {
        initProgressBar();
        initBackToTop();
        initCollapsibleAdvantages();
        initActiveToCLinks();
        initSmoothScrollToC();
    }

    // Run on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
