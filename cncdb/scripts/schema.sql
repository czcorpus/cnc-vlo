CREATE TABLE vlo_metadata_corpus (
  id int(11) PRIMARY KEY NOT NULL AUTO_INCREMENT,
  corpus_name varchar(63) NOT NULL,
  CONSTRAINT vlo_metadata_corpus_corpus_name_fk FOREIGN KEY (corpus_name) REFERENCES kontext_corpus(name)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE vlo_metadata_service (
  id int(11) PRIMARY KEY NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  description varchar(255) NOT NULL,
  link varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE vlo_metadata_common (
  id int(11) PRIMARY KEY NOT NULL AUTO_INCREMENT,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  deleted TINYINT(1) DEFAULT 0,
  type ENUM('corpus', 'service') NOT NULL,
  title_en VARCHAR(255) NOT NULL,
  title_cs VARCHAR(255) NOT NULL,
  license_info VARCHAR(255) NOT NULL,
  contact_user_id INT(11) NOT NULL,
  authors TEXT NOT NULL,
  corpus_metadata_id INT,
  service_metadata_id INT,
  CONSTRAINT vlo_metadata_common_contact_user_id_fk FOREIGN KEY (contact_user_id) REFERENCES kontext_user(id),
  CONSTRAINT vlo_metadata_common_corpus_metadata_id_fk FOREIGN KEY (corpus_metadata_id) REFERENCES vlo_metadata_corpus(id),
  CONSTRAINT vlo_metadata_common_service_metadata_id_fk FOREIGN KEY (service_metadata_id) REFERENCES vlo_metadata_service(id)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;