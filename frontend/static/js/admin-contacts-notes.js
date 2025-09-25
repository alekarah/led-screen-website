// Общая логика вкладки «Заметки» и «Перезвонить позже» для обеих модалок.
// Использование:
//   const Notes = ContactsNotes.init({
//     getCurrentId: () => currentId,
//     els: { tabInfoBtn, tabNotesBtn, tabInfo, tabNotes, remAt, remSave, remClear, notesList, noteAuthor, noteText, noteAdd }
//   });
//   // при открытии модалки:
//   Notes.onOpen({ remindAt: btn.dataset.remindAt || '' });

(function (w) {
  function init({ getCurrentId, els }) {
    const f = els || {};

    // ——— Табы ———
    function showTab(tab) {
      const isInfo = tab === "info";
      f.tabInfo?.classList.toggle("hidden", !isInfo);
      f.tabNotes?.classList.toggle("hidden", isInfo);
      f.tabInfoBtn?.classList.toggle("btn-blue", isInfo);
      f.tabNotesBtn?.classList.toggle("btn-blue", !isInfo);
    }
    f.tabInfoBtn?.addEventListener("click", () => showTab("info"));
    f.tabNotesBtn?.addEventListener("click", () => showTab("notes"));

    // ——— Заметки ———
    async function loadNotesSafe() {
      const id = getCurrentId?.();
      if (!f.notesList || !id) return;
      try {
        const { notes } = await w.ContactsAPI.getNotes(id);
        renderNotes(notes || []);
      } catch {
        renderNotes([]);
      }
    }

    function renderNotes(notes) {
      if (!f.notesList) return;
      f.notesList.innerHTML = "";
      if (!notes.length) {
        const li = document.createElement("li");
        li.style.color = "#777";
        li.textContent = "Заметок пока нет";
        f.notesList.appendChild(li);
        return;
      }
      for (const n of notes) {
        const li = document.createElement("li");
        li.dataset.noteId = n.id;
        const dateStr = n.created_at ? new Date(n.created_at).toLocaleString("ru-RU") : "";
        const author = n.author ? ` — ${n.author}` : "";
        li.innerHTML = `<span>${escapeHtml(n.text)}</span><span style="color:#777;"> (${dateStr}${author})</span>
                        <button type="button" class="btn btn-small js-note-del" style="margin-left:6px;">Удалить</button>`;
        f.notesList.appendChild(li);
      }
    }

    function escapeHtml(s) {
      return String(s || "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", '"':"&quot;", "'":"&#39;" }[ch]));
    }

    f.noteAdd?.addEventListener("click", async () => {
      const id = getCurrentId?.();
      if (!id) return;
      const text = (f.noteText?.value || "").trim();
      const author = (f.noteAuthor?.value || "").trim();
      if (!text) { w.ContactsUI.show("error", "Введите текст заметки"); return; }
      try {
        await w.ContactsAPI.addNote(id, { text, author });
        f.noteText.value = "";
        await loadNotesSafe();
        w.ContactsUI.show("ok", "Заметка добавлена");
      } catch (err) {
        w.ContactsUI.show("error", err.message);
      }
    });

    f.notesList?.addEventListener("click", async (e) => {
      const btn = e.target.closest(".js-note-del");
      if (!btn) return;
      const id = getCurrentId?.();
      if (!id) return;
      const li = btn.closest("li");
      const noteId = Number(li?.dataset.noteId || 0);
      if (!noteId) return;
      if (!confirm("Удалить заметку?")) return;
      try {
        await w.ContactsAPI.deleteNote(id, noteId);
        li.remove();
        if (!f.notesList.children.length) renderNotes([]);
        w.ContactsUI.show("ok", "Заметка удалена");
      } catch (err) {
        w.ContactsUI.show("error", err.message);
      }
    });

    // ——— Напоминание ———
    f.remSave?.addEventListener("click", async () => {
      const id = getCurrentId?.();
      if (!id || !f.remAt) return;
      const v = (f.remAt.value || "").trim(); // "YYYY-MM-DDTHH:MM"
      if (!v) { w.ContactsUI.show("error", "Выберите дату и время"); return; }
      try {
        await w.ContactsAPI.setReminder(id, { remind_at: v.replace("T", " ") });
        w.ContactsUI.show("ok", "Напоминание сохранено");
      } catch (err) {
        w.ContactsUI.show("error", err.message);
      }
    });

    f.remClear?.addEventListener("click", async () => {
      const id = getCurrentId?.();
      if (!id || !f.remAt) return;
      try {
        await w.ContactsAPI.setReminder(id, { remind_at: "" });
        f.remAt.value = "";
        w.ContactsUI.show("ok", "Напоминание очищено");
      } catch (err) {
        w.ContactsUI.show("error", err.message);
      }
    });

    // ——— Публичный API модуля ———
    async function onOpen({ remindAt } = {}) {
      if (f.remAt) f.remAt.value = remindAt || "";
      await loadNotesSafe();
      showTab("info");
    }

    return { onOpen, showTab, reloadNotes: loadNotesSafe };
  }

  w.ContactsNotes = { init };
})(window);
