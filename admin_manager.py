#!/usr/bin/env python3
import psycopg2
import bcrypt
import os

# Database connection
def get_db_connection():
    if os.getenv('LOCAL_DB') == 'true':
        return psycopg2.connect(
            host="localhost",
            port=5432,
            database="swadiq",
            user="postgres",
            password=""
        )
    else:
        return psycopg2.connect(
            host="129.80.85.203",
            port=5432,
            database="swadiq",
            user="imaad",
            password="Ertdfgxc",
            connect_timeout=10
        )

def hash_password(password):
    return bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')

def setup_tables():
    """Create missing tables if they don't exist"""
    conn = get_db_connection()
    cur = conn.cursor()
    
    try:
        # Enable UUID extension
        cur.execute('CREATE EXTENSION IF NOT EXISTS "uuid-ossp"')
        
        # Create roles table
        cur.execute("""
            CREATE TABLE IF NOT EXISTS roles (
                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                name VARCHAR(100) UNIQUE NOT NULL,
                is_active BOOLEAN DEFAULT true,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                deleted_at TIMESTAMP WITH TIME ZONE
            )
        """)
        
        # Create user_roles table
        cur.execute("""
            CREATE TABLE IF NOT EXISTS user_roles (
                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                deleted_at TIMESTAMP WITH TIME ZONE,
                UNIQUE (user_id, role_id)
            )
        """)
        
        # Insert default roles
        cur.execute("""
            INSERT INTO roles (name) VALUES 
            ('admin'), ('head_teacher'), ('class_teacher'), ('subject_teacher')
            ON CONFLICT (name) DO NOTHING
        """)
        
        conn.commit()
        print("✓ Database tables ready")
        
    except Exception as e:
        print(f"✗ Setup error: {e}")
        # Try alternative approach without UUID extension
        try:
            cur.execute("ROLLBACK")
            print("Trying alternative setup...")
            
            # Create roles table with serial ID
            cur.execute("""
                CREATE TABLE IF NOT EXISTS roles (
                    id SERIAL PRIMARY KEY,
                    name VARCHAR(100) UNIQUE NOT NULL,
                    is_active BOOLEAN DEFAULT true,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
                )
            """)
            
            # Create user_roles table with serial ID
            cur.execute("""
                CREATE TABLE IF NOT EXISTS user_roles (
                    id SERIAL PRIMARY KEY,
                    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                    UNIQUE (user_id, role_id)
                )
            """)
            
            # Insert default roles
            cur.execute("""
                INSERT INTO roles (name) VALUES 
                ('admin'), ('head_teacher'), ('class_teacher'), ('subject_teacher')
                ON CONFLICT (name) DO NOTHING
            """)
            
            conn.commit()
            print("✓ Database tables ready (alternative setup)")
            
        except Exception as e2:
            print(f"✗ Alternative setup also failed: {e2}")
    finally:
        cur.close()
        conn.close()

def add_admin_user(email, password, first_name, last_name):
    conn = get_db_connection()
    cur = conn.cursor()
    
    try:
        hashed_password = hash_password(password)
        
        cur.execute("""
            INSERT INTO users (email, password, first_name, last_name, is_active)
            VALUES (%s, %s, %s, %s, %s)
            RETURNING id
        """, (email, hashed_password, first_name, last_name, True))
        
        user_id = cur.fetchone()[0]
        
        # Get admin role ID
        cur.execute("SELECT id FROM roles WHERE name = 'admin'")
        admin_role_id = cur.fetchone()[0]
        
        # Assign admin role
        cur.execute("""
            INSERT INTO user_roles (user_id, role_id)
            VALUES (%s, %s)
        """, (user_id, admin_role_id))
        
        conn.commit()
        print(f"✓ Admin user {email} created successfully!")
        
    except psycopg2.IntegrityError:
        print(f"✗ User {email} already exists!")
    except Exception as e:
        print(f"✗ Error: {e}")
    finally:
        cur.close()
        conn.close()

def list_users():
    conn = get_db_connection()
    cur = conn.cursor()
    
    try:
        cur.execute("""
            SELECT u.id, u.email, u.first_name, u.last_name, 
                   COALESCE(string_agg(r.name, ', '), 'No roles') as roles
            FROM users u
            LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.deleted_at IS NULL
            LEFT JOIN roles r ON ur.role_id = r.id AND r.deleted_at IS NULL
            WHERE u.deleted_at IS NULL
            GROUP BY u.id, u.email, u.first_name, u.last_name
            ORDER BY u.email
        """)
        
        users = cur.fetchall()
        
        print("\n" + "="*70)
        print("USERS LIST")
        print("="*70)
        for i, (user_id, email, first_name, last_name, roles) in enumerate(users, 1):
            print(f"{i:2d}. {email:<35} ({first_name} {last_name})")
            print(f"    Roles: {roles}")
        print("="*70)
        
        return users
        
    except Exception as e:
        print(f"✗ Error listing users: {e}")
        return []
    finally:
        cur.close()
        conn.close()

def assign_admin_role(user_id):
    conn = get_db_connection()
    cur = conn.cursor()
    
    try:
        # Get admin role ID
        cur.execute("SELECT id FROM roles WHERE name = 'admin'")
        admin_role_id = cur.fetchone()[0]
        
        # Assign admin role
        cur.execute("""
            INSERT INTO user_roles (user_id, role_id)
            VALUES (%s, %s)
            ON CONFLICT (user_id, role_id) DO NOTHING
        """, (user_id, admin_role_id))
        
        conn.commit()
        print("✓ Admin role assigned successfully!")
        
    except Exception as e:
        print(f"✗ Error: {e}")
    finally:
        cur.close()
        conn.close()

def test_connection():
    """Test database connection"""
    try:
        conn = get_db_connection()
        cur = conn.cursor()
        cur.execute("SELECT version()")
        version = cur.fetchone()[0]
        print(f"✓ Connected to: {version}")
        cur.close()
        conn.close()
        return True
    except Exception as e:
        print(f"✗ Connection failed: {e}")
        return False

def main():
    print("🏫 SWADIQ SCHOOLS ADMIN MANAGER")
    print("="*40)
    
    # Test connection first
    if not test_connection():
        print("\nTry setting LOCAL_DB=true if you have a local PostgreSQL:")
        print("export LOCAL_DB=true")
        return
    
    # Setup tables
    setup_tables()
    
    while True:
        print("\nOptions:")
        print("1. Add new admin user")
        print("2. List all users")
        print("3. Assign admin role to existing user")
        print("4. Test connection")
        print("5. Exit")
        
        choice = input("\nSelect option (1-5): ").strip()
        
        if choice == "1":
            print("\n--- Add New Admin User ---")
            email = input("Email: ").strip()
            password = input("Password: ").strip()
            first_name = input("First Name: ").strip()
            last_name = input("Last Name: ").strip()
            add_admin_user(email, password, first_name, last_name)
            
        elif choice == "2":
            list_users()
            
        elif choice == "3":
            users = list_users()
            if users:
                try:
                    user_num = int(input(f"\nSelect user number (1-{len(users)}): ")) - 1
                    if 0 <= user_num < len(users):
                        user_id = users[user_num][0]
                        email = users[user_num][1]
                        confirm = input(f"Make {email} an admin? (y/N): ").lower()
                        if confirm == 'y':
                            assign_admin_role(user_id)
                    else:
                        print("✗ Invalid selection!")
                except ValueError:
                    print("✗ Invalid input!")
            else:
                print("No users found!")
                
        elif choice == "4":
            test_connection()
            
        elif choice == "5":
            print("Goodbye! 👋")
            break
        else:
            print("✗ Invalid option!")

if __name__ == "__main__":
    main()
