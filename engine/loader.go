package engine

import (
	"bufio"
	"fmt"
	//"image"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/der-antikeks/gisp/math"
)

// TODO: caching

func LoadObject(obj, mtl string) (Object, error) {
	// load materials
	materials := map[string]Material{}
	if mtl != "" {
		var err error
		materials, err = loadMTL(mtl)
		if err != nil {
			return nil, err
		}
	}

	// open object file, init reader
	file, err := os.Open(obj)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	basePath := filepath.Dir(obj) + string(filepath.Separator)

	// root object
	group := NewGroup()
	object := group

	// starting mesh
	geo := NewGeometry()
	mat, err := NewMaterial("phong")
	if err != nil {
		return nil, err
	}
	mesh := NewMesh(geo, mat)

	// cache
	var (
		vertices []math.Vector
		normals  []math.Vector
		uvs      []math.Vector
		color    = math.Color{1, 1, 1}
	)

	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fields := strings.Split(strings.TrimSpace(line), " ")
			value := strings.TrimSpace(strings.Join(fields[1:], " "))

			switch strings.ToLower(fields[0]) {

			case "v": // vertex: x, y, z
				// v 9.035167 173.402832 -2.713000
				x, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				y, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				z, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return nil, err
				}

				vertices = append(vertices, math.Vector{x, y, z})

			case "vt": // texture: u, v
				// vt 0.748573 0.750412
				u, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				v, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				uvs = append(uvs, math.Vector{u, 1.0 - v})

			case "vn": // normal: x, y, z
				// vn 0.000000 0.000000 -1.000000
				x, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				y, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				z, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return nil, err
				}

				normals = append(normals, math.Vector{x, y, z})

			case "f": // face
				// f 3 8 4 - vertex
				// f 1/4 2/5 3/6 - vertex/uv
				// f 24//24 25//24 13//24 - vertex//normal
				// f 5/1/1 1/2/1 4/3/1 - vertex/uv/normal
				// 8/11/7 - 8 vertex index (1 based), 11 texture index, 7 normal index

				// quad instead of tri, split up
				// f 5/1/1 1/2/1 4/3/1 3/4/2
				seg := [][]string{}
				if len(fields) == 5 {
					seg = append(seg, []string{"f", fields[1], fields[2], fields[4]})
					seg = append(seg, []string{"f", fields[2], fields[3], fields[4]})
				} else {
					seg = append(seg, fields)
				}

				for _, fields := range seg {
					var face [3]Vertex
					var v uint64

					for i, f := range fields[1:4] {
						a := strings.Split(f, "/")
						face[i].color = color

						// vertex
						if v, err = strconv.ParseUint(a[0], 10, 64); err != nil {
							return nil, err
						}
						face[i].position = vertices[v-1]

						// uv
						if len(a) > 1 && a[1] != "" {
							if v, err = strconv.ParseUint(a[1], 10, 64); err != nil {
								return nil, err
							}
							face[i].uv = uvs[v-1]
						}

						// normal
						if len(a) == 3 {
							if v, err = strconv.ParseUint(a[2], 10, 64); err != nil {
								return nil, err
							}
							face[i].normal = normals[v-1]
						}
					}

					geo.AddFace(face[0], face[1], face[2])
				}

			case "o": // new object
				object = NewGroup()
				group.AddChild(object)

			case "g": // mesh within object
				if geo.VerticesCount() > 0 {
					geo.MergeVertices()
					geo.ComputeBoundary()

					object.AddChild(mesh)

					geo = NewGeometry()
					mesh = NewMesh(geo, mat)
				}

			case "usemtl": // material name for the element following it
				if m, ok := materials[value]; ok {
					mat = &m
					mesh.SetMaterial(mat)
				} else {
					mat, err = NewMaterial("phong")
					if err != nil {
						return nil, err
					}
				}

			case "mtllib": // mtl file
				if mtl == "" {
					if materials, err = loadMTL(basePath + value); err != nil {
						//return nil, err
						fmt.Println("could not load mtl file:", err.Error())
					}
				}

			case "s": // smooth shading
			case "#": // comment
			case "": // empty line
			default:
				return nil, fmt.Errorf("unknown object line type: %s", line)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}

	// close last mesh
	if geo.VerticesCount() > 0 {
		geo.MergeVertices()
		geo.ComputeBoundary()

		object.AddChild(mesh)
	}

	return group, nil
}

func loadMTL(path string) (map[string]Material, error) {
	// open file, init reader
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	type info struct {
		name       string
		ka, kd, ks *math.Color // ambient, diffuse, specular color
		ns         float64     // specular exponent
		ni         float64     // optical density
		d          float64     // dissolve
		illum      int         // illumination model
		map_kd     string      // diffuse texture map
	}

	basePath := filepath.Dir(path) + string(filepath.Separator)

	// cache
	var tmp *info
	materials := map[string]*info{}

	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fields := strings.Split(strings.TrimSpace(line), " ")
			value := strings.TrimSpace(strings.Join(fields[1:], " "))

			switch strings.ToLower(fields[0]) {
			case "newmtl": // name
				tmp = &info{
					name: value,
				}
				materials[value] = tmp

			case "ns": // specular exponent
				v, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}
				tmp.ns = v

			case "ka": // ambient color
				r, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				g, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				b, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return nil, err
				}

				c := math.ColorFromRGB(r, g, b)
				tmp.ka = &c

			case "kd": // diffuse color
				r, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				g, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				b, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return nil, err
				}

				c := math.ColorFromRGB(r, g, b)
				tmp.kd = &c

			case "ks": // specular color
				r, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}

				g, err := strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return nil, err
				}

				b, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return nil, err
				}

				c := math.ColorFromRGB(r, g, b)
				tmp.ks = &c

			case "ni": // optical density
				v, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}
				tmp.ni = v

			case "d": // disolve
				v, err := strconv.ParseFloat(fields[1], 64)
				if err != nil {
					return nil, err
				}
				tmp.d = v

			case "illum": // illumination model
				v, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return nil, err
				}
				tmp.illum = int(v)

			case "map_kd": // diffuse texture map
				tmp.map_kd = basePath + value

			case "#": // comment
			case "": // empty line
			default:
				return nil, fmt.Errorf("unknown material line type: %s", line)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}

	results := map[string]Material{}

	for n, i := range materials {
		// invert transparency
		//i.d = 1 - i.d

		m, err := NewMaterial("phong")
		if err != nil {
			return nil, err
		}

		// diffuse
		if i.kd != nil {
			//m.SetDiffuseColor(*i.kd)
			m.SetUniform("diffuse", *i.kd)
		}

		// ambient
		if i.ka != nil {
			//m.SetAmbientColor(*i.ka)
			m.SetUniform("ambient", *i.ka)
		} else if i.kd != nil {
			//m.SetAmbientColor(m.DiffuseColor())
			m.SetUniform("ambient", m.Uniform("ambient"))
		}

		// specular
		if i.ks != nil {
			//m.SetSpecularColor(*i.ks)
			m.SetUniform("specular", *i.ks)
		}

		// diffuse texture map
		if i.map_kd != "" {
			tx, err := LoadTexture(i.map_kd)
			if err != nil {
				return nil, err
			}

			//m.SetDiffuseMap(tx)
			m.SetUniform("diffuseMap", tx)
		}

		//m.SetShininess(i.ns) // specular exponent
		m.SetUniform("shininess", i.ns)
		//i.ni // optical density
		//m.SetOpacity(i.d) // dissolve
		m.SetUniform("opacity", i.d)
		//i.illum // illumination model

		results[n] = *m
	}

	return results, nil
}
