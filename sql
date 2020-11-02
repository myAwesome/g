CREATE DATABASE `school` /*!40100 DEFAULT CHARACTER SET latin1 */;

CREATE TABLE `discipline` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=latin1;

CREATE TABLE `lesson` (
  `id` int(11) NOT NULL,
  `date` date DEFAULT NULL,
  `discipline_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_lesson_1_idx` (`discipline_id`),
  CONSTRAINT `fk_lesson_1` FOREIGN KEY (`discipline_id`) REFERENCES `discipline` (`id`) ON DELETE NO ACTION ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE `student` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `firstname` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=latin1;

CREATE TABLE `student_on_lesson` (
  `student_id` int(11) NOT NULL,
  `lesson_id` int(11) NOT NULL,
  KEY `fk_student_on_lesson_1_idx` (`student_id`),
  KEY `fk_student_on_lesson_2_idx` (`lesson_id`),
  CONSTRAINT `fk_student_on_lesson_1` FOREIGN KEY (`student_id`) REFERENCES `student` (`id`) ON DELETE NO ACTION ON UPDATE NO ACTION,
  CONSTRAINT `fk_student_on_lesson_2` FOREIGN KEY (`lesson_id`) REFERENCES `lesson` (`id`) ON DELETE NO ACTION ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
