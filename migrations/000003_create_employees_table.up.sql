create table Employees (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  departament_id INT NOT NULL,
  project_id INT NOT NULL,
  FOREIGN KEY(departament_id) REFERENCES departaments (id) ON DELETE CASCADE ON UPDATE cascade,
  FOREIGN KEY(project_id) REFERENCES Projects (id) ON DELETE CASCADE ON UPDATE CASCADE
)
