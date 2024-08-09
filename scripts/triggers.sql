--
-- This is a general version of triggers for synchronizing
-- descriptions between the `corpora` and `vlo_metadata_common` tables.
-- It should work with any non-cnc instance of KonText in case
-- the installation is based on mysql_* plugins.
--


DELIMITER //

DROP TRIGGER IF EXISTS sync_descriptions_from_corpora_trig //
DROP TRIGGER IF EXISTS sync_descriptions_from_metadata_trig //

CREATE TRIGGER sync_descriptions_from_corpora_trig
AFTER UPDATE ON kontext_corpus
FOR EACH ROW
BEGIN
    SET @contact_user_id = 1;
    IF NOT (NEW.description_cs <=> OLD.description_cs) OR NOT (NEW.description_en <=> OLD.description_en) THEN
        SELECT name INTO @corpus_name FROM kontext_corpus WHERE id = NEW.id;
        SELECT id INTO @corpus_metadata_id FROM vlo_metadata_corpus WHERE corpus_name = @corpus_name;
        IF @corpus_metadata_id IS NULL THEN
            INSERT INTO vlo_metadata_corpus (corpus_name) VALUES (@corpus_name);
            INSERT INTO vlo_metadata_common (desc_cs, desc_en, corpus_metadata_id, contact_user_id, deleted)
                VALUES (COALESCE(NEW.description_cs,''), COALESCE(NEW.description_en,''), LAST_INSERT_ID(), @contact_user_id, 1);
        ELSEIF @skip_vlo_update IS NULL THEN
            SET @skip_corpora_update = 1;
            UPDATE vlo_metadata_common SET desc_cs = COALESCE(NEW.description_cs,''), desc_en = COALESCE(NEW.description_en,'') WHERE corpus_metadata_id = @corpus_metadata_id;
            SET @skip_corpora_update = NULL;
        END IF;
    END IF;
END;
//

CREATE TRIGGER sync_descriptions_from_metadata_trig
AFTER UPDATE ON vlo_metadata_common
FOR EACH ROW
BEGIN
    IF NEW.type = 'corpus' AND (NOT (NEW.desc_cs <=> OLD.desc_cs) OR NOT (NEW.desc_en <=> OLD.desc_en)) THEN
        IF @skip_corpora_update IS NULL THEN
            SET @skip_vlo_update = 1;
            UPDATE kontext_corpus SET description_cs = NULLIF(NEW.desc_cs,''), description_en = NULLIF(NEW.desc_en,'') WHERE name = (SELECT corpus_name FROM vlo_metadata_corpus WHERE id = NEW.corpus_metadata_id);
            SET @skip_vlo_update = NULL;
        END IF;
    END IF;
END;
//